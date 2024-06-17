package tsengine

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/direktiv/direktiv/pkg/runtime"
	"github.com/dop251/goja"
	"github.com/google/uuid"
)

type RuntimeManager struct {
	baseFS string

	// flowInformation *compiler.FlowInformation

	mtx sync.Mutex
}

type Status struct {
	Start  int64 `json:"start"`
	Active int32 `json:"active"`
}

const (
	managerFsShared    = "shared"
	managerFsInstances = "instances"
)

const (
	StateDataInputFile = "input.data"
)

func New(baseFS string) (*RuntimeManager, error) {

	manager := &RuntimeManager{
		baseFS: baseFS,
	}

	// prepare filesystem
	err := os.MkdirAll(filepath.Join(baseFS, managerFsShared), 0766)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(filepath.Join(baseFS, managerFsInstances), 0766)
	if err != nil {
		return nil, err
	}

	return manager, nil
}

func (rm *RuntimeManager) NewHandler(prg *goja.Program, fn string, secrets map[string]string,
	functions map[string]string, jsonInput bool) RuntimeHandler {
	rm.mtx.Lock()
	defer rm.mtx.Unlock()

	return RuntimeHandler{
		secrets:     secrets,
		prg:         prg,
		jsonPayload: jsonInput,
		startFn:     fn,
		functions:   functions,
	}
}

type RuntimeHandler struct {
	baseFS string

	secrets, functions map[string]string
	startFn            string

	prg         *goja.Program
	jsonPayload bool

	// flowInformation *compiler.FlowInformation
	Status Status

	mtx *sync.Mutex
}

var _ http.Handler = RuntimeHandler{}

func (rh RuntimeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	fmt.Println("RUN REQUEST!!!!")

	id := uuid.New()

	atomic.AddInt32(&rh.Status.Active, 1)
	defer atomic.AddInt32(&rh.Status.Active, -1)

	instanceDir := filepath.Join(rh.baseFS, "instances", id.String())
	err := os.MkdirAll(instanceDir, 0755)
	if err != nil {
		writeError(w, direktivErrorInternal, err.Error())
		return
	}
	defer os.RemoveAll(instanceDir)

	rt, err := runtime.New(id, rh.prg, &rh.secrets, &rh.functions, rh.baseFS, rh.jsonPayload)
	if err != nil {
		writeError(w, direktivErrorInternal, err.Error())
		return
	}

	ret, state, err := rt.Execute(rh.startFn, r, w)
	if err != nil {
		writeError(w, direktivErrorInternal, err.Error())
		return
	}

	// only write return if not directly written to response
	if !state.Response().Written && ret != nil {
		r, err := json.Marshal(ret)
		if err != nil {
			writeError(w, direktivErrorInternal, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(r)
	}

}
