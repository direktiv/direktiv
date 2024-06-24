package tsengine

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/direktiv/direktiv/pkg/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Status struct {
	Start int64 `json:"start"`
}

const (
	StateDataInputFile = "input.data"
)

func NewHandler(cfg Config, db *database.SQLStore) (RuntimeHandler, error) {
	handler := RuntimeHandler{
		baseDir: cfg.BaseDir,
		db:      db,
		ctx: engineCtx{
			Namespace:     cfg.Namespace,
			WorkflowsPath: cfg.WorkflowPath,
		},
	}

	return handler, nil
}

type RuntimeHandler struct {
	baseDir string
	db      *database.SQLStore
	mtx     *sync.Mutex
	ctx     engineCtx
}

// engineCtx defines the context in which
// the engine operates (e.g., namespace and workflow paths)
// its main purpose is logging & tracing.
type engineCtx struct {
	Namespace     string
	WorkflowsPath string
	// TODO more attr
}

var _ http.Handler = RuntimeHandler{}

func (rh RuntimeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rh.mtx.Lock()
	defer rh.mtx.Unlock()
	// TODO compile the program so its ready to be served
	// TODO actual execution of the program
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
	config := Config{}

	if err := env.Parse(config); err != nil {
		return nil, fmt.Errorf("parsing env variables: %w", err)
	}
	slog.Info("initializing the database")

	db, err := initDB(config)
	if err != nil {
		return nil, fmt.Errorf("was unable create db %w", err)
	}
	slog.Info("database initialized")

	handler, err := NewHandler(config, db)
	if err != nil {
		return nil, fmt.Errorf("was unable to create handler %w", err)
	}

	// handle flow requests
	s.mux.HandleFunc("/", handler.ServeHTTP)
	s.mux.HandleFunc("GET /status", s.HandleStatusRequest)

	// TODO: cancel
	// s.mux.HandleFunc("GET /cancel/{id}", s.HandleStatusRequest)
	slog.Info("started engine server")

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

func initDB(config Config) (*database.SQLStore, error) {
	// TODO: this should be re-done
	gormConf := &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				LogLevel:                  logger.Silent,
				IgnoreRecordNotFoundError: true,
			},
		),
	}

	var err error
	var db *gorm.DB
	//nolint:intrange
	for i := 0; i < 10; i++ {
		slog.Info("connecting to database...")

		db, err = gorm.Open(postgres.New(postgres.Config{
			DSN:                  config.DBConfig,
			PreferSimpleProtocol: false, // disables implicit prepared statement usage
			// Conn:                 edb.SQLStore(),
		}), gormConf)
		if err == nil {
			slog.Info("successfully connected to the database.")

			break
		}
		time.Sleep(time.Second)
	}

	if err != nil {
		return nil, err
	}

	gdb, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("modifying gorm driver, err: %w", err)
	}

	slog.Debug("Database connection pool limits set", "maxIdleConns", 32, "maxOpenConns", 16)
	gdb.SetMaxIdleConns(32)
	gdb.SetMaxOpenConns(16)

	dbManager := database.NewSQLStore(db, config.SecretKey)

	return dbManager, nil
}
