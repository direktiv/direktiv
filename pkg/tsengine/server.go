package tsengine

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/tsengine/tsservice"
	"github.com/direktiv/direktiv/pkg/tsengine/tstypes"
	"github.com/dop251/goja"
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

// NewHandler creates and initializes a RuntimeHandler for a TypeScript service.
func NewHandler(serverCtx context.Context, cfg Config, db *database.SQLStore) (RuntimeHandler, error) {
	handler := RuntimeHandler{
		baseDir: cfg.BaseDir,
		db:      db,
		ctx: WorkflowContext{
			Namespace:    cfg.Namespace,
			WorkflowPath: cfg.WorkflowPath,
		},
	}

	var err error

	// Fetch the TypeScript file from the database.
	file, err := handler.db.FileStore().ForNamespace(handler.ctx.Namespace).
		GetFile(serverCtx, handler.ctx.WorkflowPath)
	if err != nil {
		return handler, fmt.Errorf("failed to retrieve TypeScript file: %w", err)
	}

	file.Data, err = handler.db.FileStore().ForFile(file).GetData(serverCtx)
	if err != nil {
		return handler, fmt.Errorf("failed to get file data for '%v': %w", file.Name(), err)
	}

	// Compile the TypeScript code.
	compiler, err := tsservice.NewTSServiceCompiler(handler.ctx.Namespace, handler.ctx.WorkflowPath, string(file.Data))
	if err != nil {
		return handler, fmt.Errorf("TypeScript compilation error: %w", err)
	}

	handler.execCtx, err = compiler.Parse()
	if err != nil {
		return handler, fmt.Errorf("failed to parse TypeScript code: %w", err)
	}

	handler.prog, err = compiler.Compile(goja.CompileAST)
	if err != nil {
		return handler, fmt.Errorf("failed to finalize TypeScript compilation: %w", err)
	}

	// TODO: ... (boot-up Go Functions as cmds for the vm later) ...

	return handler, nil
}

type RuntimeHandler struct {
	baseDir string
	db      *database.SQLStore
	ctx     WorkflowContext
	execCtx *tstypes.ExecutionContext
	prog    *goja.Program
}

// WorkflowContext defines the context in which
// the engine operates (e.g., namespace and workflow paths)
// its main purpose is logging & tracing.
type WorkflowContext struct {
	Namespace    string
	WorkflowPath string
	// TODO more attr
}

var _ http.Handler = RuntimeHandler{}

// ServeHTTP handles the HTTP request for the TypeScript service.
func (rh RuntimeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rt := goja.New()

	// TODO: ... (Injecting Go Functions) ...

	gojaRes, err := rt.RunProgram(rh.prog)
	if err != nil {
		slog.Error("Failed to execute TypeScript program", "error", err, "workflowPath", rh.ctx.WorkflowPath)
		http.Error(w, "failed to execute program", http.StatusInternalServerError)

		return
	}

	if err := writeResultResponse(w, gojaRes.Export()); err != nil {
		slog.Error("Failed to write response", "error", err, "workflowPath", rh.ctx.WorkflowPath)
		http.Error(w, "failed to write response", http.StatusInternalServerError)
	}
}

func writeResultResponse(w http.ResponseWriter, result interface{}) error {
	w.Header().Set("Content-Type", "application/json")

	switch res := result.(type) {
	case nil:
		result = "" // Handle nil as an empty string

	case string, float64, int, map[string]interface{}, []interface{}:
		// These types can be directly marshaled
		// No additional conversion needed

	default:
		// Handle unsupported types or try conversion to string
		var ok bool
		if result, ok = convertResultToString(res); !ok {
			return fmt.Errorf("unsupported result type: %T", result)
		}
	}
	writeJSONResponse(w, map[string]interface{}{"result": result})

	return nil
}

// TODO: evaluate if need this
// Helper function to attempt string conversion for unsupported types.
func convertResultToString(result interface{}) (any, bool) {
	// TODO: Add custom logic here to attempt converting our types to JSON-marshallable types
	// For example, you could use fmt.Sprintf("%v", result) as a general fallback

	// TODO add our types here if we really need that
	// if res, ok := result.("My-direktiv-custom types"); ok {
	// 	return res.String(), true
	// }

	// TODO: Fallback (might not always be appropriate)
	return fmt.Sprintf("%v", result), true
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

	if err := env.Parse(&config); err != nil {
		return nil, fmt.Errorf("parsing env variables: %w", err)
	}
	slog.Info("initializing the database")

	db, err := initDB(config)
	if err != nil {
		return nil, fmt.Errorf("was unable create db %w", err)
	}
	slog.Info("database initialized")

	handler, err := NewHandler(context.Background(), config, db)
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

func writeJSONResponse(w http.ResponseWriter, data interface{}) {
	jsonResponse, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "failed to marshal response", http.StatusInternalServerError)
		return
	}
	_, err = w.Write(jsonResponse)
	if err != nil {
		http.Error(w, "failed to write response", http.StatusInternalServerError)
		return
	}
}
