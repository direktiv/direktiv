package tsengine

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/direktiv/direktiv/pkg/runtime"
	"github.com/google/uuid"
)

type Engine struct {
	baseFS string

	// initData InitData

	Status Status

	mtx sync.Mutex

	manager runtime.Manager
}

type Status struct {
	Start  int64 `json:"start"`
	Active int32 `json:"active"`
}

const (
	engineFsShared    = "shared"
	engineFsInstances = "instances"
)

func New(baseFS string, manager runtime.Manager) (*Engine, error) {
	engine := &Engine{
		baseFS: baseFS,
		Status: Status{
			Start: time.Now().UnixMilli(),
		},
		manager: manager,
	}

	// prepare filesystem
	err := os.MkdirAll(filepath.Join(baseFS, engineFsShared), 0766)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(filepath.Join(baseFS, engineFsInstances), 0766)
	if err != nil {
		return nil, err
	}

	return engine, nil
}

func (e *Engine) RunRequest(req *http.Request, resp http.ResponseWriter) {
	id := uuid.New()

	atomic.AddInt32(&e.Status.Active, 1)
	defer atomic.AddInt32(&e.Status.Active, -1)

	instanceDir := filepath.Join(e.baseFS, "instances", id.String())
	err := os.MkdirAll(instanceDir, 0755)
	if err != nil {
		writeError(resp, direktivErrorInternal, err.Error())
		return
	}
	defer os.RemoveAll(instanceDir)

	rt, err := runtime.New(id, e.baseFS, e.manager)
	if err != nil {
		writeError(resp, direktivErrorInternal, err.Error())
		return
	}

	ret, state, err := rt.Execute(e.manager.RuntimeData().Definition.State,
		req, resp)
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
