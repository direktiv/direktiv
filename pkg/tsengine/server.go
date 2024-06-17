package tsengine

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"gorm.io/gorm"
)

type Server struct {
	mux *http.ServeMux
	srv *http.Server

	Engine *Engine
	// initializer Initializer
}

const (
	ServerInitDB   = "db"
	ServerInitFile = "file"
)

func NewServer(cfg Config, db *gorm.DB) (*Server, error) {
	slog.Info("starting engine server")
	slog.Info(fmt.Sprintf("using flow %s", cfg.FlowPath))
	slog.Info(fmt.Sprintf("using base dir %s", cfg.BaseDir))

	// get path
	mux := &http.ServeMux{}
	s := &Server{
		mux: mux,
		srv: &http.Server{Addr: ":8080", Handler: mux},
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

	var enviroment Environment
	if db != nil {
		enviroment = NewDBEnviroment(cfg.BaseDir, cfg.FlowPath, cfg.Namespace, cfg.SecretKey, db, engine)
	} else {
		fi := NewFileEnviroment(cfg.BaseDir, cfg.FlowPath, engine)
		go fi.fileWatcher(cfg.FlowPath)
		enviroment = fi
	}

	err = enviroment.Init()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Server) Start() error {
	slog.Info("starting engine")
	return s.srv.ListenAndServe()
}

func (s *Server) HandleFlowRequest(w http.ResponseWriter, r *http.Request) {
	s.Engine.RunRequest(r, w)
}

func (s *Server) HandleStatusRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.Engine.Status)
}
