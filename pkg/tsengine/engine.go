package tsengine

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/direktiv/direktiv/pkg/runtime"
	"github.com/dop251/goja"
	"github.com/google/uuid"
)

type Engine struct {
	baseFS string

	secrets, functions map[string]string
	startFn            string

	prg         *goja.Program
	jsonPayload bool

	// flowInformation *compiler.FlowInformation
	Status Status

	mtx sync.Mutex
}

type Status struct {
	Initialized bool  `json:"initialized"`
	Start       int64 `json:"start"`
	Failed      bool  `json:"failed"`
	Active      int32 `json:"active"`
}

const (
	engineFsShared    = "shared"
	engineFsInstances = "instances"
)

func New(baseFS string) (*Engine, error) {

	engine := &Engine{
		secrets: make(map[string]string),
		baseFS:  baseFS,
		Status: Status{
			Start: time.Now().UnixMilli(),
		},
	}

	var err error
	defer func() {
		if err != nil {
			engine.Status.Failed = true
		}
	}()

	// prepare filesystem
	err = os.MkdirAll(filepath.Join(baseFS, engineFsShared), 0766)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(filepath.Join(baseFS, engineFsInstances), 0766)
	if err != nil {
		return nil, err
	}

	return engine, nil
}

func (e *Engine) Initialize(prg *goja.Program, fn string, secrets map[string]string,
	functions map[string]string, jsonInput bool) {
	e.mtx.Lock()
	defer e.mtx.Unlock()

	e.secrets = secrets
	e.prg = prg
	e.jsonPayload = jsonInput
	e.startFn = fn
	e.functions = functions

	// set ok
	e.Status.Initialized = true
	e.Status.Failed = false
}

// SetError marks the engine for garbage collection if init failed
func (e *Engine) SetError(err error) {
	slog.Error("engine configuration failed", slog.Any("error", err))
	fmt.Println(e)
	fmt.Println(e.Status)
	e.Status.Initialized = true
	e.Status.Failed = true
}

func (e *Engine) RunRequest(req *http.Request, resp http.ResponseWriter) {
	id := uuid.New()

	if !e.Status.Initialized {
		writeError(resp, direktivErrorInternal, "engine not initialized")
		return
	}

	atomic.AddInt32(&e.Status.Active, 1)
	defer atomic.AddInt32(&e.Status.Active, -1)

	instanceDir := filepath.Join(e.baseFS, "instances", id.String())
	err := os.MkdirAll(instanceDir, 0755)
	if err != nil {
		writeError(resp, direktivErrorInternal, err.Error())
		return
	}
	defer os.RemoveAll(instanceDir)

	rt, err := runtime.New(id, e.prg, &e.secrets, &e.functions, e.baseFS, e.jsonPayload)
	if err != nil {
		writeError(resp, direktivErrorInternal, err.Error())
		return
	}

	ret, state, err := rt.Execute(e.startFn, req, resp)
	if err != nil {
		writeError(resp, direktivErrorInternal, err.Error())
		return
	}

	// only write return if not directly written to response
	if !state.Response().Written && ret != nil {
		r, err := json.Marshal(ret)
		if err != nil {
			writeError(resp, direktivErrorInternal, err.Error())
			return
		}
		resp.Header().Set("Content-Type", "application/json")
		resp.Write(r)
	}

}

const (
	StateDataInputFile = "input.data"
)
