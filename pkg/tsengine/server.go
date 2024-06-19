package tsengine

import (
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/dop251/goja"
)

type RuntimeManager struct {
	baseFS string
	mtx    sync.Mutex
}

type Status struct {
	Start int64 `json:"start"`
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
	// prepare filesystem.
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
		baseFS:      rm.baseFS,
	}
}

type RuntimeHandler struct {
	baseFS      string
	secrets     map[string]string
	functions   map[string]string
	startFn     string
	prg         *goja.Program
	jsonPayload bool
	Status      Status
}

var _ http.Handler = RuntimeHandler{}

func (rh RuntimeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {}

func GenerateBasicServiceFile(path, ns string) *core.ServiceFileData {
	return &core.ServiceFileData{
		Typ:       core.ServiceTypeTypescript,
		Name:      path,
		Namespace: ns,
		FilePath:  path,
	}
}
