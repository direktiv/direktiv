package tsengine

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

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

type Server struct {
	mux *http.ServeMux
	srv *http.Server

	Handler http.Handler
	Status  Status
}

func NewServer() (*Server, error) {
	slog.Info("starting engine server")

	mux := &http.ServeMux{}
	s := &Server{
		mux: mux,
		srv: &http.Server{
			Addr:              ":8080",
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
		},
		Status: Status{
			Start: time.Now().UnixMilli(),
		},
	}

	// handle flow requests
	s.mux.HandleFunc("/", http.NotFound)

	s.mux.HandleFunc("GET /status", s.HandleStatusRequest)

	// TODO: cancel
	// s.mux.HandleFunc("GET /cancel/{id}", s.HandleStatusRequest)

	return s, nil
}

func (s *Server) HandleStatusRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(s.Status); err != nil {
		slog.Error("Failed to encode status response", slog.Any("error", err))
		http.Error(w, "Failed to encode status response", http.StatusInternalServerError)
	}
}

func (s *Server) Start() error {
	slog.Info("Starting engine")
	err := s.srv.ListenAndServe()
	if err != nil {
		slog.Error("Server encountered an error", slog.Any("error", err))
		return fmt.Errorf("server encountered an error: %w", err)
	}

	return nil
}
