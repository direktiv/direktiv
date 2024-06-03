package tsengine

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

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

	// copy itself to shared location
	if cfg.SelfCopy != "" {
		ex, err := os.Executable()
		if err != nil {
			return nil, err
		}

		slog.Info("copying binary", slog.String("source", ex),
			slog.String("target", cfg.SelfCopy))
		_, err = copyFile(ex, cfg.SelfCopy)
		if err != nil {
			panic(err)
		}

		err = os.Chmod(cfg.SelfCopy, 0755)
		if err != nil {
			panic(err)
		}
	}

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

	var initializer Initializer
	if db != nil {
		initializer = NewDBInitializer(cfg.BaseDir, cfg.FlowPath, cfg.Namespace, db, engine)
	} else {
		fi := NewFileInitializer(cfg.BaseDir, cfg.FlowPath, engine)
		go fi.fileWatcher(cfg.FlowPath)
		initializer = fi
	}

	err = initializer.Init()
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
