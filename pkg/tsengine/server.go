package tsengine

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/direktiv/direktiv/pkg/datastore/datastoresql"
	"github.com/direktiv/direktiv/pkg/filestore/filestoresql"
	"github.com/direktiv/direktiv/pkg/tsengine/environment"

	"gorm.io/gorm"
)

type Server struct {
	mux *http.ServeMux
	srv *http.Server

	Handler http.Handler
	Status  Status

	// initializer Initializer
}

const (
	ServerInitDB   = "db"
	ServerInitFile = "file"
)

var _ http.Handler = Server{}

func NewServer(cfg Config, db *gorm.DB) (*Server, error) {
	slog.Info("starting engine server")
	slog.Info(fmt.Sprintf("using flow %s", cfg.FlowPath))
	slog.Info(fmt.Sprintf("using base dir %s", cfg.BaseDir))

	// get path
	mux := &http.ServeMux{}
	s := &Server{
		mux: mux,
		srv: &http.Server{Addr: ":8080", Handler: mux},
		Status: Status{
			Start: time.Now().UnixMilli(),
		},
	}
	var err error
	s.Handler, err = CreateRuntimeHandler(cfg, db)
	if err != nil {
		return nil, err
	}

	// handle flow requests
	s.mux.HandleFunc("/", s.ServeHTTP)

	s.mux.HandleFunc("GET /status", s.HandleStatusRequest)

	// TODO: cancel
	// s.mux.HandleFunc("GET /cancel/{id}", s.HandleStatusRequest)

	return s, nil
}

func CreateRuntimeHandler(cfg Config, db *gorm.DB) (*RuntimeHandler, error) {
	slog.Info("starting engine server")
	slog.Info(fmt.Sprintf("using flow %s", cfg.FlowPath))
	slog.Info(fmt.Sprintf("using base dir %s", cfg.BaseDir))

	engine, err := New(cfg.BaseDir)
	if err != nil {
		return nil, err
	}

	driver, err := buildDriver(db, engine.baseFS, cfg)
	if err != nil {
		return nil, err
	}

	compiler, err := environment.BuildCompiler(context.Background(), driver, cfg.Namespace, cfg.FlowPath)
	if err != nil {
		return nil, err
	}

	flowInfo, err := compiler.CompileFlow()
	if err != nil {
		return nil, err
	}

	functions := environment.NewFunctionBuilder(*flowInfo, engine.baseFS).Build()
	secrets := environment.NewSecretBuilder(driver, cfg.Namespace, *flowInfo, cfg.BaseDir).Build(context.Background())
	watcher := environment.NewFileBuilder(driver, cfg.Namespace, *flowInfo, engine.baseFS).Build(context.Background())
	go watcher.Watch(context.Background(), cfg.FlowPath)
	handler := engine.NewHandler(compiler.Program, flowInfo.Definition.State, secrets, functions, flowInfo.Definition.Json)

	return &handler, nil
}

func buildDriver(db *gorm.DB, baseFS string, cfg Config) (environment.Driver, error) {
	if db != nil {
		ds := datastoresql.NewSQLStore(db, cfg.SecretKey)
		fs := filestoresql.NewSQLFileStore(db)

		return &environment.DBBasedProvider{
			SecretsStore: ds.Secrets(),
			FileStore:    fs,
			FlowFilePath: cfg.FlowPath,
			BaseFilePath: baseFS,
		}, nil
	}
	return &environment.FileBasedProvider{BaseFilePath: baseFS}, nil
}

func (s *Server) Start() error {
	slog.Info("starting engine")
	return s.srv.ListenAndServe()
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Handler.ServeHTTP(w, r)
}

func (s *Server) HandleStatusRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.Status)
}
