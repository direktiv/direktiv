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

type Runtime struct {
	// program     *goja.Program
	vm      *goja.Runtime
	id      uuid.UUID
	baseDir string
	manager Manager
	// jsonInput   bool

	// Secrets   *map[string]string
	// Functions *map[string]string
}

// type RuntimeData struct {
// 	Program   *goja.Program
// 	JSONInput bool

// 	Secrets   *map[string]string
// 	Functions *map[string]string
// }

func (rt *Runtime) Execute(fn string, req *http.Request, resp http.ResponseWriter) (interface{}, *State, error) {

	rt.manager.CreateInstance(rt.id, "api", rt.manager.RuntimeData().Script)

	defer func() {
		if r := recover(); r != nil {
			slog.Error("recover", slog.Any("panic", r))
		}
	}()

	function, ok := goja.AssertFunction(rt.vm.Get(fn))
	if !ok {
		return nil, nil, fmt.Errorf("function %s not found", fn)
	}

	if req.URL == nil {
		req.URL = &url.URL{}
	}

	// create file if input is file
	data, err := prepareInput(req, rt.dirInfo().instanceDir,
		rt.manager.RuntimeData().Definition.Json)
	if !ok {
		return nil, nil, err
	}

	state := NewState(resp, data, req.Header, req.URL.Query(), rt)
	value, err := function(goja.Undefined(), rt.vm.ToValue(state))

	// "unwrap" error if it is one of ours
	if err != nil {
		var ge *goja.Exception
		if errors.As(err, &ge) {
			e := err.(*goja.Exception).Value().Export()
			errMap, ok := e.(map[string]interface{})
			if ok {
				code, ok := errMap["code"]
				if !ok {
					code = DirektivErrorCode
				}
				msg, ok := errMap["msg"]
				if !ok {
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

func New(id uuid.UUID, baseDir string, manager Manager) (*Runtime, error) {

	slog.Debug("creating new runtime", slog.String("dir", baseDir), slog.String("instance", id.String()))

	vm := goja.New()
	vm.SetMaxCallStackSize(25)
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
	vm.SetParserOptions(parser.WithDisableSourceMaps)

	// static commands
	vm.Set("log", commands.Log)
	vm.Set("sleep", commands.Sleep)
	vm.Set("atob", commands.Atob)
	vm.Set("btoa", commands.Btoa)
	vm.Set("toJSON", commands.ToJSON)
	vm.Set("fromJSON", commands.FromJSON)
	vm.Set("trim", commands.Trim)

	rt := &Runtime{
		vm:      vm,
		id:      id,
		baseDir: baseDir,

		manager: manager,
	}

	vm.Set("getSecret", rt.getSecret)
	vm.Set("getFile", rt.getFile)
	vm.Set("setupFunction", rt.setupFunction)
	vm.Set("httpRequest", rt.HttpRequest)

	if rt.manager.RuntimeData().Program == nil {
		return nil, fmt.Errorf("no program provided")
	}

	// run this to prepare for
	_, err := rt.vm.RunProgram(rt.manager.RuntimeData().Program)
	if err != nil {
		return nil, err
	}

	return rt, nil
}

type dirInfo struct {
	sharedDir, instanceDir string
}

func (rt *Runtime) dirInfo() *dirInfo {
	return &dirInfo{
		sharedDir:   filepath.Join(rt.baseDir, SharedDir),
		instanceDir: filepath.Join(rt.baseDir, InstancesDir, rt.id.String()),
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
		// handle json and other types
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
			return data, err
		} else {
			return string(b), nil
		}
	}

	// if configured in flow, we save it as file
	p := filepath.Join(instanceDir, StateDataInputFile)
	out, err := os.Create(p)
	if err != nil {
		return nil, err
	}
	defer out.Close()
	_, err = io.Copy(out, req.Body)
	return nil, err
}
