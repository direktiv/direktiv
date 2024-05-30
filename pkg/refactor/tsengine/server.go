package tsengine

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v11"
)

type Server struct {
	mux *http.ServeMux
	srv *http.Server

	Engine      *Engine
	initializer Initializer
	prefix      string
}

const (
	ServerInitMux  = "mux"
	ServerInitFile = "file"
)

func NewServer() (*Server, error) {

	// parsing config
	cfg := Config{}
	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}
	// utils.SetLogLevel(cfg.LogLevel)

	slog.Info("starting engine server")
	slog.Info(fmt.Sprintf("using %s initializer", cfg.Initializer))
	slog.Info(fmt.Sprintf("using flow %s", cfg.FlowPath))
	slog.Info(fmt.Sprintf("using base dir %s", cfg.BaseDir))

	// copy itself to shared location
	if cfg.SelfCopy != "" {
		slog.Info("copying binary")
		_, err = copyFile("/engine", cfg.SelfCopy)
		if err != nil {
			panic(err)
		}

		err := os.Chmod(cfg.SelfCopy, 0777)
		if err != nil {
			panic(err)
		}
	}

	// get path
	mux := &http.ServeMux{}
	s := &Server{
		mux:    mux,
		srv:    &http.Server{Addr: ":8080", Handler: mux},
		prefix: "myprefix",
	}

	engine, err := New(cfg.BaseDir)
	if err != nil {
		return nil, err
	}
	s.Engine = engine

	// handle flow requests
	s.mux.HandleFunc("/", s.HandleFlowRequest)

	s.mux.HandleFunc("GET /status", s.HandleStatusRequest)

	// TODO: cancel
	// s.mux.HandleFunc("GET /cancel/{id}", s.HandleStatusRequest)

	var initializer Initializer
	switch cfg.Initializer {
	case ServerInitMux:
		initializer = NewMuxInitializer(s.prefix, filepath.Join(cfg.BaseDir, engineFsShared), mux, engine)
	case ServerInitFile:
		fi := NewFileInitializer(cfg.BaseDir, cfg.FlowPath, engine)
		go fi.fileWatcher(cfg.FlowPath)
		initializer = fi
	default:
		return nil, fmt.Errorf("unknown initializer")
	}

	s.initializer = initializer

	return s, nil
}

func (s *Server) Prefix() string {
	return s.prefix
}

func (s *Server) Initializer() Initializer {
	return s.initializer
}

func (s *Server) Start() error {
	slog.Info("starting engine")
	go s.initializer.Init()
	return s.srv.ListenAndServe()
}

func (s *Server) HandleFlowRequest(w http.ResponseWriter, r *http.Request) {
	if !s.Engine.Status.Initialized {
		w.Write([]byte("not initialized"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	s.Engine.RunRequest(r, w)
}

func (s *Server) HandleStatusRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.Engine.Status)
}
