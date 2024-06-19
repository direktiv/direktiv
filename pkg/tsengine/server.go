package tsengine

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/engine"
	"github.com/dop251/goja"
)

type RuntimeManager struct {
	baseFS string
	mtx    sync.Mutex
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
	baseFS      string
	secrets     map[string]string
	functions   map[string]string
	startFn     string
	tracingAttr engine.ActionContext
	prg         *goja.Program
	jsonPayload bool
	Status      Status
}

var _ http.Handler = RuntimeHandler{}

func (rh RuntimeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("RUN REQUEST!!!!")
	atomic.AddInt32(&rh.Status.Active, 1)
	defer atomic.AddInt32(&rh.Status.Active, -1)
}

func GenerateBasicServiceFile(path, ns string) *core.ServiceFileData {
	return &core.ServiceFileData{
		Typ:       core.ServiceTypeTypescript,
		Name:      path,
		Namespace: ns,
		FilePath:  path,
	}
}
