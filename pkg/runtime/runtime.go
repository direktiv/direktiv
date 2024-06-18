package runtime

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/direktiv/direktiv/pkg/runtime/commands"
	"github.com/dop251/goja"
	"github.com/dop251/goja/parser"
	"github.com/google/uuid"
)

// Runtime represents a runtime environment for executing JavaScript code.
type Runtime struct {
	vm        *goja.Runtime
	id        string
	baseDir   string
	jsonInput bool
}

// Executor handles the execution of JavaScript functions within a runtime environment.
type Executor struct {
	rt *Runtime
}

// Execute runs the specified function within the runtime environment.
func (e *Executor) Execute(fn string, req *http.Request, resp http.ResponseWriter) (interface{}, *State, error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("recover", slog.Any("panic", r))
		}
	}()

	function, ok := goja.AssertFunction(e.rt.vm.Get(fn))
	if !ok {
		return nil, nil, fmt.Errorf("function %s not found", fn)
	}

	if req.URL == nil {
		req.URL = &url.URL{}
	}

	// Create file if input is file
	data, err := prepareInput(req, e.rt.dirInfo().instanceDir, e.rt.jsonInput)
	if err != nil {
		return nil, nil, err
	}

	state := NewState(resp, data, req.Header, req.URL.Query(), e.rt)
	value, err := function(goja.Undefined(), e.rt.vm.ToValue(state))

	if err != nil {
		var ge *goja.Exception
		if errors.As(err, &ge) {
			e := ge.Value().Export()
			errMap, ok := e.(map[string]interface{})
			if ok {
				code, _ := errMap["code"]
				if code == "" {
					code = DirektivErrorCode
				}
				msg, _ := errMap["msg"]
				if msg == "" {
					msg = err
				}
				return nil, nil, NewDirektivError(code, msg)
			}
		}
		return nil, nil, err
	}

	var retValue interface{}
	if !state.response.Written {
		retValue = value.Export()
	}

	return retValue, state, nil
}

func (rt *Runtime) Prepare(prg *goja.Program) (*Executor, error) {
	if prg == nil {
		return nil, fmt.Errorf("no program provided")
	}

	_, err := rt.vm.RunProgram(prg)
	if err != nil {
		return nil, err
	}
	return &Executor{
		rt: rt,
	}, nil
}

func New(id uuid.UUID, baseDir string, jsonInput bool) (*Runtime, error) {
	slog.Debug("creating new runtime", slog.String("dir", baseDir), slog.String("instance", id.String()))

	vm := goja.New()
	vm.SetMaxCallStackSize(25)
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
	vm.SetParserOptions(parser.WithDisableSourceMaps)

	rt := &Runtime{
		vm:        vm,
		id:        id.String(),
		baseDir:   baseDir,
		jsonInput: jsonInput,
	}

	err := vm.Set("log", commands.Log)
	if err != nil {
		return nil, err
	}
	err = vm.Set("sleep", commands.Sleep)
	if err != nil {
		return nil, err
	}
	err = vm.Set("atob", commands.Atob)
	if err != nil {
		return nil, err
	}
	err = vm.Set("btoa", commands.Btoa)
	if err != nil {
		return nil, err
	}
	err = vm.Set("toJSON", commands.ToJSON)
	if err != nil {
		return nil, err
	}
	err = vm.Set("fromJSON", commands.FromJSON)
	if err != nil {
		return nil, err
	}
	err = vm.Set("trim", commands.Trim)
	if err != nil {
		return nil, err
	}

	return rt, nil
}

type Command interface {
	GetName() string
	GetCommandFunction() interface{}
}

func (rt *Runtime) WithCommand(command Command) error {
	return rt.vm.Set(command.GetName(), command.GetCommandFunction())
}

type dirInfo struct {
	sharedDir, instanceDir string
}

func (rt *Runtime) dirInfo() *dirInfo {
	return &dirInfo{
		sharedDir:   filepath.Join(rt.baseDir, SharedDir),
		instanceDir: filepath.Join(rt.baseDir, InstancesDir, rt.id),
	}
}

func prepareInput(req *http.Request, instanceDir string, asJSON bool) (interface{}, error) {
	if req.Body == nil {
		req.Body = io.NopCloser(strings.NewReader(""))
	}
	defer req.Body.Close()

	if req.URL == nil {
		req.URL = &url.URL{}
	}

	if asJSON {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}

		if json.Valid(b) {
			var data interface{}
			err = json.Unmarshal(b, &data)
			if err != nil {
				return nil, err
			}
			return data, nil
		}
		return string(b), nil
	}

	p := filepath.Join(instanceDir, StateDataInputFile)
	out, err := os.Create(p)
	if err != nil {
		return nil, err
	}
	defer out.Close()
	_, err = io.Copy(out, req.Body)
	return nil, err
}
