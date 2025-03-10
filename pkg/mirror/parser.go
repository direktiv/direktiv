package mirror

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/telemetry"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

type Parser struct {
	matcher gitignore.Matcher
	src     Source
	tempDir string

	Filters   map[string][]byte
	Workflows map[string][]byte
	Services  map[string][]byte
	Endpoints map[string][]byte
	Consumers map[string][]byte
	Gateways  map[string][]byte

	DeprecatedNamespaceVars map[string][]byte
	DeprecatedWorkflowVars  map[string]map[string][]byte

	Namespace, PID string
}

func NewParser(namespace, pid string, src Source) (*Parser, error) {
	tempDir, err := os.MkdirTemp("", "direktiv_sync_*")
	if err != nil {
		return nil, err
	}

	telemetry.LogActivityDebug(fmt.Sprintf("processing repository in temporary directory: %s", tempDir),
		namespace, pid)

	p := &Parser{
		matcher: gitignore.NewMatcher(nil),
		src:     src,
		tempDir: tempDir,

		Filters:   make(map[string][]byte),
		Workflows: make(map[string][]byte),
		Services:  make(map[string][]byte),
		Endpoints: make(map[string][]byte),
		Consumers: make(map[string][]byte),
		Gateways:  make(map[string][]byte),

		DeprecatedNamespaceVars: make(map[string][]byte),
		DeprecatedWorkflowVars:  make(map[string]map[string][]byte),

		Namespace: namespace,
		PID:       pid,
	}

	err = p.parse()
	if err != nil {
		telemetry.LogActivityError("error processing repository",
			namespace, pid, err)
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
		telemetry.LogActivityInfo("no .direktivignore file detected", p.Namespace, p.PID)

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
			telemetry.LogActivityInfo(fmt.Sprintf("skipping directory '%s': excluded by .direktivignore patterns", path),
				p.Namespace, p.PID)

			return fs.SkipDir
		}

		telemetry.LogActivityInfo(fmt.Sprintf("skipping file '%s': excluded by .direktivignore patterns", path),
			p.Namespace, p.PID)

		return nil
	}

	base := filepath.Base(path)
	_, err := filestore.SanitizePath(base)
	if err != nil {
		if d.IsDir() {
			telemetry.LogActivityInfo(fmt.Sprintf("skipping directory '%s': filename contains illegal characters", path),
				p.Namespace, p.PID)

			return fs.SkipDir
		}

		telemetry.LogActivityInfo(fmt.Sprintf("skipping file '%s': filename contains illegal characters", path),
			p.Namespace, p.PID)

		return nil
	}

	tpath := filepath.Join(p.tempDir, path)

	if d.IsDir() {
		err := os.MkdirAll(tpath, perms)
		if err != nil {
			return err
		}

		telemetry.LogActivityDebug(fmt.Sprintf("created directory '%s'", path),
			p.Namespace, p.PID)

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

	telemetry.LogActivityDebug(fmt.Sprintf("created file '%s'", path),
		p.Namespace, p.PID)

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
		telemetry.LogActivityWarn(fmt.Sprintf("error loading possible Direktiv resource definition '%s': %v", path, err),
			p.Namespace, p.PID)

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
	case *core.ConsumerFile:
		err = p.handleConsumer(path, data)
		if err != nil {
			return err
		}
	case *core.ServiceFile:
		err = p.handleService(path, data)
		if err != nil {
			return err
		}
	case core.Endpoint:
		err = p.handleEndpoint(path, data)
		if err != nil {
			return err
		}
	case core.Gateway:
		err = p.handleGateway(path, data)
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

	telemetry.LogActivityDebug(fmt.Sprintf("pruned Direktiv resource file '%s'", path),
		p.Namespace, p.PID)

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
		telemetry.LogActivityWarn(fmt.Sprintf("error loading possible Direktiv workflow definition (ambiguous) '%s': %v", path, err),
			p.Namespace, p.PID)

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

	telemetry.LogActivityDebug(fmt.Sprintf("pruned Direktiv workflow definition file '%s'", path),
		p.Namespace, p.PID)

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
	telemetry.LogActivityInfo(fmt.Sprintf("direktiv resource file containing a workflow definition found at '%s'", path),
		p.Namespace, p.PID)

	p.Workflows[path] = data

	return nil
}

func (p *Parser) handleService(path string, data []byte) error {
	telemetry.LogActivityInfo(fmt.Sprintf("direktiv resource file containing a service definition found at '%s'", path),
		p.Namespace, p.PID)

	p.Services[path] = data

	return nil
}

func (p *Parser) handleGateway(path string, data []byte) error {
	telemetry.LogActivityInfo(fmt.Sprintf("direktiv resource file containing a gateway definition found at '%s'", path),
		p.Namespace, p.PID)

	p.Gateways[path] = data

	return nil
}

func (p *Parser) handleEndpoint(path string, data []byte) error {
	telemetry.LogActivityInfo(fmt.Sprintf("direktiv resource file containing an endpoint definition found at '%s'", path),
		p.Namespace, p.PID)

	p.Endpoints[path] = data

	return nil
}

func (p *Parser) handleConsumer(path string, data []byte) error {
	telemetry.LogActivityInfo(fmt.Sprintf("direktiv resource file containing a consumer definition found at '%s'", path),
		p.Namespace, p.PID)

	p.Consumers[path] = data

	return nil
}

func (p *Parser) handleFilters(path string, filters *model.Filters) error {
	telemetry.LogActivityInfo(fmt.Sprintf("direktiv resource file containing %d filter definitions found at '%s'",
		len(filters.Filters), path),
		p.Namespace, p.PID)

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
		telemetry.LogActivityInfo(fmt.Sprintf("filter '%s' loaded", filter.Name),
			p.Namespace, p.PID)
	}

	return nil
}

func (p *Parser) logUnprunedFiles() error {
	tfs := os.DirFS(p.tempDir)

	err := fs.WalkDir(tfs, ".", func(path string, d fs.DirEntry, err error) error {
		if !d.Type().IsRegular() {
			return nil
		}

		telemetry.LogActivityInfo(fmt.Sprintf("file '%s' loaded", path),
			p.Namespace, p.PID)

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
	regex := regexp.MustCompile(datastore.RuntimeVariableNameRegexPattern)

	allFiles, err := p.listOnlyFiles()
	if err != nil {
		return err
	}

	for _, fpath := range allFiles {
		base := filepath.Base(fpath)
		prefix := "var."
		vname := strings.TrimPrefix(base, prefix)

		if !regex.MatchString(vname) {
			telemetry.LogActivityWarn(fmt.Sprintf("detected a possible deprecated namespace variable definition with an invalid name at: %s", fpath),
				p.Namespace, p.PID)

			continue
		}

		if strings.HasPrefix(base, prefix) {
			actual := filepath.Join(p.tempDir, fpath)

			data, err := os.ReadFile(actual)
			if err != nil {
				return err
			}

			p.DeprecatedNamespaceVars[vname] = data

			telemetry.LogActivityWarn(fmt.Sprintf("detected deprecated namespace variable definition at: %s", fpath),
				p.Namespace, p.PID)
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
				telemetry.LogActivityWarn(fmt.Sprintf("detected a possible deprecated workflow variable definition with an invalid name at: %s", fpath),
					p.Namespace, p.PID)

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

				telemetry.LogActivityWarn(fmt.Sprintf("detected deprecated workflow variable definition at: %s", fpath),
					p.Namespace, p.PID)
			}
		}
	}

	return nil
}
