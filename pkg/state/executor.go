package state

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

	"github.com/direktiv/direktiv/pkg/runtime"
	"github.com/dop251/goja"
)

// Executor handles the execution of JavaScript functions within a runtime environment.
type Executor struct {
	rt *runtime.Runtime
}

// Execute runs the specified function within the runtime environment.
func (e *Executor) Execute(fn string, req *http.Request, resp http.ResponseWriter) (interface{}, *State, error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("recover", slog.Any("panic", r))
		}
	}()

	function, ok := goja.AssertFunction(e.rt.VM.Get(fn))
	if !ok {
		return nil, nil, fmt.Errorf("function %s not found", fn)
	}

	if req.URL == nil {
		req.URL = &url.URL{}
	}

	// Create file if input is file
	data, err := prepareInput(req, e.rt.DirInfo().InstanceDir, e.rt.JsonInput)
	if err != nil {
		return nil, nil, err
	}

	state := NewState(resp, data, req.Header, req.URL.Query(), e.rt)
	value, err := function(goja.Undefined(), e.rt.VM.ToValue(state))

	if err != nil {
		var ge *goja.Exception
		if errors.As(err, &ge) {
			e := err.(*goja.Exception).Value().Export()
			errMap, ok := e.(map[string]interface{})
			if ok {
				code, ok := errMap["code"]
				if !ok {
					code = runtime.DirektivErrorCode
				}
				msg, ok := errMap["msg"]
				if !ok {
					msg = err
				}
				return nil, nil, runtime.NewDirektivError(code, msg)
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

func New(rt *runtime.Runtime, prg *goja.Program) (*Executor, error) {
	if prg == nil {
		return nil, fmt.Errorf("no program provided")
	}

	_, err := rt.VM.RunProgram(prg)
	if err != nil {
		return nil, err
	}
	return &Executor{
		rt: rt,
	}, nil
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

	p := filepath.Join(instanceDir, runtime.StateDataInputFile)
	out, err := os.Create(p)
	if err != nil {
		return nil, err
	}
	defer out.Close()
	_, err = io.Copy(out, req.Body)
	return nil, err
}
