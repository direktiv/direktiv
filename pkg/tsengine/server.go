package tsengine

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	"github.com/dop251/goja/ast"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Status struct {
	Start int64 `json:"start"`
}

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
	// store the compiler func for compiling a program per request.
	handler.compile = compiler.Compile

	return handler, nil
}

type RuntimeHandler struct {
	baseDir string
	db      *database.SQLStore
	ctx     WorkflowContext
	execCtx *tstypes.ExecutionContext
	compile func(compileAst func(prg *ast.Program, strict bool) (*goja.Program, error)) (*goja.Program, error)
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
	prog, err := rh.compile(goja.CompileAST)
	if err != nil {
		logAndHTTPError(w, "Failed to compile TypeScript to program", err, rh.ctx)
		return
	}

	input, err := io.ReadAll(r.Body)
	if err != nil {
		logAndHTTPError(w, "Failed to read input", err, rh.ctx)
		return
	}

	// Handle Input based on Content Type and Sources.
	if err := rh.handleInput(rt, r, input); err != nil {
		logAndHTTPError(w, "Failed to handle input", err, rh.ctx)
		return
	}

	// TODO: ... (Injecting Go Functions) ...

	// Run the main TypeScript program
	_, err = rt.RunProgram(prog) // discard results we expect them to be undefined / nil
	if err != nil {
		logAndHTTPError(w, "Failed to execute TypeScript program (main)", err, rh.ctx)
		return
	}

	slog.Info("executing prog", "state", rh.execCtx.Definition.State)

	// Prepare and execute the state-specific code
	var args []goja.Value
	if rt.Get("input") != goja.Undefined() {
		args = append(args, rt.Get("input"))
	}

	stateCall := fmt.Sprintf("%s(%s)", rh.execCtx.Definition.State, marshalArgsToJSON(args))
	result, err := rt.RunString(stateCall)

	if err != nil {
		logAndHTTPError(w, "Failed to execute TypeScript program (state)", err, rh.ctx)
		return
	}

	slog.Info("got results", "res", result)
	if err := writeResultResponse(w, result.Export()); err != nil {
		logAndHTTPError(w, "Failed to write response", err, rh.ctx)
		return
	}
}

// handleInput manages the input based on the content type and available sources.
func (rh RuntimeHandler) handleInput(rt *goja.Runtime, r *http.Request, input []byte) error {
	if len(input) > 0 && r.Header.Get("Content-Type") == "application/json" {
		var inputData map[string]interface{}
		if err := json.Unmarshal(input, &inputData); err != nil {
			return err
		}

		return rt.Set("input", inputData)
	} else if len(input) > 0 {
		return rt.Set("input", string(input))
	} else if len(rh.execCtx.Input) > 0 {
		return rt.Set("input", rh.execCtx.Input)
	}

	// No input available
	return nil
}

func logAndHTTPError(w http.ResponseWriter, msg string, err error, ctx WorkflowContext) {
	// TODO: use direktiv error format.
	slog.Error(msg, "error", err, "workflowPath", ctx.WorkflowPath)
	http.Error(w, msg, http.StatusInternalServerError)
}

func writeResultResponse(w http.ResponseWriter, result interface{}) error {
	w.Header().Set("Content-Type", "application/json")

	switch res := result.(type) {
	case nil:
		result = "" // TODO should we handle nil as an empty string?

	case string, float64, int, map[string]interface{}, []interface{}:
		// These types can be directly marshaled.
		// No additional conversion needed.

	default:
		// Handle unsupported types or try conversion to string.
		var ok bool
		if result, ok = convertResultToString(res); !ok {
			return fmt.Errorf("unsupported result type: %T", result)
		}
	}
	writeJSONResponse(w, map[string]interface{}{"result": result})

	return nil
}

// TODO: evaluate if need this.
// Helper function to attempt string conversion for unsupported types.
func convertResultToString(result interface{}) (any, bool) {
	// TODO: Add custom logic here to attempt converting our types to JSON-marshallable types.
	// For example, you could use fmt.Sprintf("%v", result) as a general fallback.

	// TODO add our types here if we really need that
	// if res, ok := result.("My-direktiv-custom types"); ok {
	// 	return res.String(), true
	// }

	// TODO: Fallback (might not always be appropriate).
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

// marshalArgsToJSON converts a slice of Goja values to a JSON string for use as function arguments.
func marshalArgsToJSON(args []goja.Value) string {
	var goArgs []interface{}
	for _, arg := range args {
		goArgs = append(goArgs, arg.Export())
	}
	jsonArgs, _ := json.Marshal(goArgs)
	return string(jsonArgs)
}
