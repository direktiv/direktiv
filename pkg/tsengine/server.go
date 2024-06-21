package tsengine

import (
	"net/http"
	"sync"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/dop251/goja"
)

type RuntimeHandlerBuilder struct {
	baseFS string
	mtx    sync.Mutex
}

type Status struct {
	Start int64 `json:"start"`
}

const (
	StateDataInputFile = "input.data"
)

func NewBuilder(baseFS string) (*RuntimeHandlerBuilder, error) {
	manager := &RuntimeHandlerBuilder{
		baseFS: baseFS,
	}

	return manager, nil
}

func (rm *RuntimeHandlerBuilder) NewHandler(prg *goja.Program, fn string, secrets map[string]string, functions map[string]string, jsonInput bool) RuntimeHandler {
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
