package tsengine

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/compiler"
	"github.com/fsnotify/fsnotify"
)

type Initializer interface {
	Init()
}

type FileInitializer struct {
	srcDir, flowPath string
	engine           *Engine
}

func NewFileInitializer(srcDir, flowPath string, e *Engine) *FileInitializer {
	return &FileInitializer{
		srcDir:   srcDir,
		flowPath: flowPath,
		engine:   e,
	}
}

func (i *FileInitializer) Init() {
	slog.Info("reading flow")

	b, err := os.ReadFile(i.flowPath)
	if err != nil {
		i.engine.SetError(err)
		return
	}

	c, err := compiler.New(i.flowPath, string(b))
	if err != nil {
		i.engine.SetError(err)
		return
	}

	fi, err := c.CompileFlow()
	if err != nil {
		i.engine.SetError(err)
		return
	}

	secrets := make(map[string]string)

	// read secrets
	for a := range fi.Secrets {
		s := fi.Secrets[a]
		secretFile := filepath.Join(i.engine.baseFS, "secrets", s.Name)
		content, err := os.ReadFile(secretFile)
		if err != nil {
			slog.Error("can not read secret", slog.String("secret", s.Name), slog.Any("error", err))
			continue
		}
		secrets[s.Name] = string(content)
	}

	// read files
	for a := range fi.Files {
		file := fi.Files[a]
		if file.Scope == "shared" {
			filePathSrc := filepath.Join(i.engine.baseFS, file.Name)
			filePathTarget := filepath.Join(i.engine.baseFS, "shared", file.Name)
			_, err := copyFile(filePathSrc, filePathTarget)
			if err != nil {
				slog.Error("can not read file", slog.String("file", file.Name), slog.Any("error", err))
				continue
			}
		}
	}

	functions := make(map[string]string)

	for i := range fi.Functions {
		f := fi.Functions[i]
		functions[f.GetID()] = os.Getenv(f.GetID())
		slog.Info("adding function", slog.String("function", f.GetID()))
	}

	// files are already there
	i.engine.Initialize(c.Program, fi.Definition.State, secrets, functions, fi.Definition.Json)
}

func (i *FileInitializer) fileWatcher(flow string) {

	// dir to watch
	dir := filepath.Dir(flow)

	// file to watch
	file := filepath.Base(flow)

	slog.Info("watching flow", slog.String("flow", flow))

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	// listening for flow changes
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					continue
				}
				if filepath.Base(event.Name) == file && event.Has(fsnotify.Write) {
					slog.Info("updating flow")
					i.Init()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					continue
				}
				slog.Error("error occurred watching flow file:", slog.Any("error", err))
			}
		}
	}()

	err = watcher.Add(dir)
	if err != nil {
		panic(err)
	}

	<-make(chan struct{})

}

type MuxInitializer struct {
	mux         *http.ServeMux
	prefix, dir string
	engine      *Engine
}

func NewMuxInitializer(prefix string, dir string, mux *http.ServeMux, e *Engine) *MuxInitializer {
	return &MuxInitializer{
		mux:    mux,
		prefix: prefix,
		dir:    dir,
		engine: e,
	}
}

func (m *MuxInitializer) Init() {
	m.mux.HandleFunc("/init", m.HandleInitRequest)
}

func (m *MuxInitializer) HandleInitRequest(w http.ResponseWriter, r *http.Request) {
	data, err := r.MultipartReader()
	if err != nil {
		m.engine.SetError(err)
		return
	}

	var script string
	var secrets = make(map[string]string)
	flowPath := os.Getenv("DIREKTIV_JSENGINE_FLOWPATH")

	for {
		part, err_part := data.NextPart()
		if err_part == io.EOF {
			break
		}

		// all requests come with a prefix to avoid name collusion
		name := part.FormName()[len(m.prefix)+1:]

		// handle script
		if name == flowPath {
			var b bytes.Buffer
			err := readPart(part, &b)
			if err != nil {
				m.engine.SetError(err)
				return
			}
			script = b.String()
			continue
		}

		split := strings.Split(name, "_")
		if len(split) != 2 {
			// should not happen because it is managed by direktiv
			slog.Warn("illegal part name", slog.String("name", name))
			continue
		}

		slog.Info("handling upload part", slog.String("part", name))

		// handle secrets
		if split[0] == "secret" {
			var b bytes.Buffer
			err := readPart(part, &b)
			if err != nil {
				m.engine.SetError(err)
				return
			}
			secrets[split[1]] = b.String()

			continue
		}

		// goes into shared directory
		if split[0] == "file" {
			f, err := os.Create(filepath.Join(m.dir, split[1]))
			if err != nil {
				m.engine.SetError(err)
				return
			}
			err = readPart(part, f)
			if err != nil {
				m.engine.SetError(err)
				return
			}
			continue
		}
	}

	// script can not be empty
	if script == "" {
		m.engine.SetError(fmt.Errorf("script can not be empty"))
		return
	}

	c, err := compiler.New(flowPath, script)
	if err != nil {
		m.engine.SetError(err)
		return
	}

	fi, err := c.CompileFlow()
	if err != nil {
		m.engine.SetError(err)
		return
	}

	functions := make(map[string]string)
	for i := range fi.Functions {
		f := fi.Functions[i]
		functions[f.GetID()] = os.Getenv(f.GetID())
	}

	m.engine.Initialize(c.Program, fi.Definition.State, secrets, functions, fi.Definition.Json)
}

func readPart(in io.Reader, out io.Writer) error {
	data := make([]byte, 1024)
	for {
		data = data[:cap(data)]
		len, err := in.Read(data)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		out.Write(data[:len])
	}

	return nil
}
