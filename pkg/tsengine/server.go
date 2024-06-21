package tsengine

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/core"
)

type Status struct {
	Start int64 `json:"start"`
}

const (
	StateDataInputFile = "input.data"
)

func NewHandler(baseFS string) (RuntimeHandler, error) {
	manager := RuntimeHandler{
		baseFS: baseFS,
	}
	// TODO Load data via configuration for the db!
	return manager, nil
}

type RuntimeHandler struct {
	baseFS string
	mtx    *sync.Mutex
}

var _ http.Handler = RuntimeHandler{}

func (rh RuntimeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rh.mtx.Lock()
	defer rh.mtx.Unlock()
	// TODO compile the program so its ready to be served
	// TODO actual execution of the program
}

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

	handler, err := NewHandler("")
	if err != nil {
		panic(err)
	}
	// handle flow requests
	s.mux.HandleFunc("/", handler.ServeHTTP)
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
