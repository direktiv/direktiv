package mirror

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/spec"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

type Parser struct {
	matcher   gitignore.Matcher
	src       Source
	tempDir   string
	namespace string
	processID string

	Filters   map[string][]byte
	Workflows map[string][]byte
	Services  map[string][]byte
	Endpoints map[string][]byte
	Consumers map[string][]byte

	DeprecatedNamespaceVars map[string][]byte
	DeprecatedWorkflowVars  map[string]map[string][]byte
}

func NewParser(namespace string, processID string, src Source) (*Parser, error) {
	tempDir, err := os.MkdirTemp("", "direktiv_sync_*")
	if err != nil {
		return nil, err
	}

	slog.Debug("Processing repository in temporary directory", "temp_path", tempDir)

	p := &Parser{
		namespace: namespace,
		processID: processID,
		matcher:   gitignore.NewMatcher(nil),
		src:       src,
		tempDir:   tempDir,

		Filters:   make(map[string][]byte),
		Workflows: make(map[string][]byte),
		Services:  make(map[string][]byte),
		Endpoints: make(map[string][]byte),
		Consumers: make(map[string][]byte),

		DeprecatedNamespaceVars: make(map[string][]byte),
		DeprecatedWorkflowVars:  make(map[string]map[string][]byte),
	}

	err = p.parse()
	if err != nil {
		slog.Error("Processing repository in temporary directory", "error", err, "namespace", p.namespace, "track", recipient.Mirror.String()+"."+p.processID)
		_ = p.Close()

		return nil, err
	}

	return p, nil
}

func (p *Parser) Close() error {
	return os.RemoveAll(p.tempDir)
}

func (p *Parser) parse() error {
	err := p.loadIgnores()
	if err != nil {
		return err
	}

	err = p.filterCopySource()
	if err != nil {
		return err
	}

	err = p.scanAndPruneDirektivResourceFiles()
	if err != nil {
		return err
	}

	err = p.scanAndPruneAmbiguousDirektivWorkflowFiles()
	if err != nil {
		return err
	}

	err = p.parseDeprecatedVariableFiles()
	if err != nil {
		return err
	}

	err = p.logUnprunedFiles()
	if err != nil {
		return err
	}

	return nil
}

func (p *Parser) loadIgnores() error {
	f, err := p.src.FS().Open(".direktivignore")
	if errors.Is(err, os.ErrNotExist) {
		slog.Debug("No .direktivignore file detected", "track", recipient.Mirror.String()+"."+p.processID)

		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to open direktivignore file: %w", err)
	}
	defer f.Close()

	var ps []gitignore.Pattern
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		s := scanner.Text()
		if !strings.HasPrefix(s, "#") && len(strings.TrimSpace(s)) > 0 {
			ps = append(ps, gitignore.ParsePattern(s, nil))
		}
	}

	err = f.Close()
	if err != nil {
		return fmt.Errorf("failed to close direktivignore file: %w", err)
	}

	p.matcher = gitignore.NewMatcher(ps)

	return nil
}

const perms = 0o755

func (p *Parser) filterCopySourceWalkFunc(path string, d fs.DirEntry, _ error) error {
	isMatch := p.matcher.Match(strings.Split(path, "/"), d.IsDir())
	if isMatch {
		if d.IsDir() {
			slog.Debug("Skipping directory, excluded by .direktivignore patterns", "path", path, "namespace", p.namespace)

			return fs.SkipDir
		}

		slog.Debug("Skipping file, excluded by .direktivignore patterns", "path", path, "namespace", p.namespace)

		return nil
	}

	base := filepath.Base(path)
	_, err := filestore.SanitizePath(base)
	if err != nil {
		if d.IsDir() {
			slog.Debug("Skipping directory filename contains illegal characters", "path", path, "namespace", p.namespace)

			return fs.SkipDir
		}

		slog.Debug("Skipping file. filename contains illegal characters", "path", path, "namespace", p.namespace)

		return nil
	}

	tpath := filepath.Join(p.tempDir, path)

	if d.IsDir() {
		err := os.MkdirAll(tpath, perms)
		if err != nil {
			return err
		}
		slog.Debug("Created directory", "path", path, "namespace", p.namespace)

		return nil
	}

	// NOTE: duplicating the file here isn't strictly necessary and could cause problems,
	// 	but large file sizes are a problem anyway.
	src, err := p.src.FS().Open(path)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(tpath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}

	err = src.Close()
	if err != nil {
		return err
	}

	err = dst.Close()
	if err != nil {
		return err
	}

	slog.Debug("Created file", "path", path, "namespace", p.namespace)

	return nil
}

func (p *Parser) filterCopySource() error {
	err := fs.WalkDir(p.src.FS(), ".", p.filterCopySourceWalkFunc)
	if err != nil {
		return err
	}

	return nil
}

func (p *Parser) listYAMLFiles() ([]string, error) {
	var paths []string

	tfs := os.DirFS(p.tempDir)

	err := fs.WalkDir(tfs, ".", func(path string, d fs.DirEntry, err error) error {
		if !d.Type().IsRegular() {
			return nil
		}

		if strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml") {
			paths = append(paths, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return paths, nil
}

func (p *Parser) scanAndPruneDirektivResourceFile(path string) error {
	tpath := filepath.Join(p.tempDir, path)

	data, err := os.ReadFile(tpath)
	if err != nil {
		return err
	}

	resource, err := model.LoadResource(data)
	if errors.Is(err, model.ErrNotDirektivAPIResource) {
		return nil
	}
	if err != nil {
		slog.Error("loading possible Direktiv resource definition", "error", err, "namespace", p.namespace, "track", recipient.Mirror.String()+"."+p.processID)

		return nil
	}

	switch typ := resource.(type) {
	case *model.Filters:
		filters, ok := resource.(*model.Filters)
		if !ok {
			panic(nil)
		}
		err = p.handleFilters(path, filters)
		if err != nil {
			return err
		}
	case *model.Workflow:
		err = p.handleWorkflow(path, data)
		if err != nil {
			return err
		}
	case *core.EndpointFile:
		err = p.handleEndpoint(path, data)
		if err != nil {
			return err
		}
	case *core.ConsumerFile:
		err = p.handleConsumer(path, data)
		if err != nil {
			return err
		}
	case *spec.ServiceFile:
		err = p.handleService(path, data)
		if err != nil {
			return err
		}
	default:
		panic(typ)
	}

	err = os.Remove(tpath)
	if err != nil {
		return err
	}
	slog.Debug("Pruned Direktiv resource file", "path", path, "namespace", p.namespace)

	return nil
}

func (p *Parser) scanAndPruneDirektivResourceFiles() error {
	paths, err := p.listYAMLFiles()
	if err != nil {
		return err
	}

	for _, path := range paths {
		err = p.scanAndPruneDirektivResourceFile(path)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Parser) scanAndPruneAmbiguousDirektivWorkflowFile(path string) error {
	tpath := filepath.Join(p.tempDir, path)

	data, err := os.ReadFile(tpath)
	if err != nil {
		return err
	}

	wf := new(model.Workflow)
	err = wf.Load(data)
	if err != nil {
		slog.Error("Error loading possible Direktiv workflow definition (ambiguous)", "path", path, "error", err, "track", recipient.Mirror.String()+"."+p.processID)

		return nil
	}

	err = p.handleWorkflow(path, data)
	if err != nil {
		return err
	}

	err = os.Remove(tpath)
	if err != nil {
		return err
	}

	slog.Debug("Pruned Direktiv workflow definition file", "path", path, "namespace", p.namespace)

	return nil
}

func (p *Parser) scanAndPruneAmbiguousDirektivWorkflowFiles() error {
	paths, err := p.listYAMLFiles()
	if err != nil {
		return err
	}

	for _, path := range paths {
		err = p.scanAndPruneAmbiguousDirektivWorkflowFile(path)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Parser) handleWorkflow(path string, data []byte) error {
	slog.Debug("Direktiv resource file containing a workflow definition found", "path", path, "namespace", p.namespace)

	p.Workflows[path] = data

	return nil
}

func (p *Parser) handleService(path string, data []byte) error {
	slog.Debug("Direktiv resource file containing a service definition found", "path", path, "namespace", p.namespace)

	p.Services[path] = data

	return nil
}

func (p *Parser) handleEndpoint(path string, data []byte) error {
	slog.Debug("Direktiv resource file containing an endpoint definition found", "path", path, "namespace", p.namespace)

	p.Endpoints[path] = data

	return nil
}

func (p *Parser) handleConsumer(path string, data []byte) error {
	slog.Debug("Direktiv resource file containing a consumer definition found", "path", path, "namespace", p.namespace)

	p.Consumers[path] = data

	return nil
}

func (p *Parser) handleFilters(path string, filters *model.Filters) error {
	slog.Debug("Direktiv resource file containing filter definitions found", "path", path, "namespace", p.namespace)

	for idx, filter := range filters.Filters {
		if _, exists := p.Filters[filter.Name]; exists {
			return fmt.Errorf("duplicate definition detected for filter '%s'", filter.Name)
		}

		var err error

		data := []byte(filters.Filters[idx].InlineJavascript)
		sourcePath := filters.Filters[idx].Source
		if sourcePath != "" {
			if !filepath.IsAbs(sourcePath) {
				sourcePath = filepath.Join(filepath.Dir(path), sourcePath)
			}

			actual := filepath.Join(p.tempDir, sourcePath)
			data, err = os.ReadFile(actual)
			if err != nil {
				return fmt.Errorf("failed to load filter source from '%s': %w", sourcePath, err)
			}
		}

		p.Filters[filter.Name] = data
		slog.Debug("Filter loaded.", "filet_name", filter.Name, "namespace", p.namespace)
	}

	return nil
}

func (p *Parser) logUnprunedFiles() error {
	tfs := os.DirFS(p.tempDir)

	err := fs.WalkDir(tfs, ".", func(path string, d fs.DirEntry, err error) error {
		if !d.Type().IsRegular() {
			return nil
		}
		slog.Debug("File loaded.", "path", path, "namespace", p.namespace)

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (p *Parser) ListFiles() ([]string, error) {
	var paths []string

	tfs := os.DirFS(p.tempDir)

	err := fs.WalkDir(tfs, ".", func(path string, d fs.DirEntry, err error) error {
		if path == "." || path == ".." {
			return nil
		}
		paths = append(paths, path)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return paths, nil
}

func (p *Parser) listOnlyFiles() ([]string, error) {
	allFiles, err := p.ListFiles()
	if err != nil {
		return nil, err
	}

	var trimmed []string
	for _, fpath := range allFiles {
		actual := filepath.Join(p.tempDir, fpath)
		fi, err := os.Stat(actual)
		if err != nil {
			return nil, err
		}

		if fi.Mode().IsRegular() {
			trimmed = append(trimmed, fpath)
		}
	}

	return trimmed, nil
}

func (p *Parser) parseDeprecatedVariableFiles() error {
	regex := regexp.MustCompile(core.RuntimeVariableNameRegexPattern)

	allFiles, err := p.listOnlyFiles()
	if err != nil {
		return err
	}

	for _, fpath := range allFiles {
		base := filepath.Base(fpath)
		prefix := "var."
		vname := strings.TrimPrefix(base, prefix)

		if !regex.MatchString(vname) {
			slog.Error("Detected a possible deprecated namespace variable definition with an invalid name", "path", fpath, "track", recipient.Mirror.String()+"."+p.processID, "namespace", p.namespace)

			continue
		}

		if strings.HasPrefix(base, prefix) {
			actual := filepath.Join(p.tempDir, fpath)

			data, err := os.ReadFile(actual)
			if err != nil {
				return err
			}

			p.DeprecatedNamespaceVars[vname] = data
			slog.Error("Detected deprecated namespace variable definition", "path", fpath, "track", recipient.Mirror.String()+"."+p.processID, "namespace", p.namespace)
		}
	}

	allWorkflows := []string{}
	for k := range p.Workflows {
		allWorkflows = append(allWorkflows, k)
	}

	for _, fpath := range allFiles {
		for _, wpath := range allWorkflows {
			prefix := wpath + "."
			vname := strings.TrimPrefix(fpath, prefix)
			if !regex.MatchString(vname) {
				slog.Error("Detected a possible deprecated workflow variable definition with an invalid name", "path", fpath, "track", recipient.Mirror.String()+"."+p.processID, "namespace", p.namespace)

				continue
			}

			if strings.HasPrefix(fpath, prefix) {
				actual := filepath.Join(p.tempDir, fpath)

				data, err := os.ReadFile(actual)
				if err != nil {
					return err
				}

				m, exists := p.DeprecatedWorkflowVars[wpath]
				if !exists {
					m = make(map[string][]byte)
					p.DeprecatedWorkflowVars[wpath] = m
				}

				m[vname] = data
				slog.Error("Detected deprecated workflow variable definition at", "path", fpath, "track", recipient.Mirror.String()+"."+p.processID, "namespace", p.namespace)
			}
		}
	}

	return nil
}
