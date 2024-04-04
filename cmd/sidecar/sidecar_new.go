package sidecar

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Config defines the configuration structure for environment variables.
type Config struct {
	InternalPort    string `env:"INTERNAL_PORT"`   // Port for the internal router.
	ExternalPort    string `env:"SIDECAR_PORT"`    // Port for the external router.
	FlowServerURL   string `env:"FLOW_SERVER_URL"` // Endpoint for forwarding task results or statuses.
	UserServiceURL  string `env:"USER_SERVICE_URL"`
	MaxResponseSize string `env:"MAX_RESPONSE_SIZE"`
}

const (
	ActionIDHeader      = "Direktiv-ActionID"
	LogLevelHeader      = "Direktiv-LogLevel"
	FilesLocationHeader = "Direktiv-TempDir"
	ErrorCodeHeader     = "Direktiv-ErrorCode"
	ErrorMessageHeader  = "Direktiv-ErrorMessage"
	ActionIDQuerryParam = "action_id"

	SharedDir = "/mnt/shared"
)

func main() {
	var config Config
	var dataMap sync.Map // actionID -> Action
	cap, err := strconv.Atoi(config.MaxResponseSize)
	if err != nil {
		config.MaxResponseSize = "134217728"
		slog.Error("Failed to read a valid value for MAX_RESPONSE_SIZE, falling back to default value")
	}
	slog.Debug("Initializing sidecar", "MaxResponseSize", cap, "FlowServerURL", config.FlowServerURL)
	// Router for handling registering new requests or canceling a ongoing one.
	// This Routes is meant to be used only ba flow
	// since we may require that the request body follows a certain type
	// If we need to expose the Service to external systems we may consider
	// Creating an other router for this usecase.
	slog.Debug("Initializing external routes")
	externalRouter := setupExternalRouter(config, &dataMap)

	// Internal router, accessable only to the user service.
	slog.Debug("Initializing external routes")
	internalRouter := setupInternalRouter(&dataMap)

	// Start routers in separate goroutines to listen on different ports.
	go func() {
		log.Fatal(http.ListenAndServe(":"+config.ExternalPort, externalRouter))
		slog.Debug("Started external routes", "port", config.ExternalPort)
	}()

	log.Fatal(http.ListenAndServe("0.0.0.0:"+config.InternalPort, internalRouter))
	slog.Debug("Started internal routes", "addr", "0.0.0.0", "port", config.InternalPort)
}

func setupExternalRouter(config Config, dataMap *sync.Map) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// Router for handling external requests.
	router.Post("/", func(w http.ResponseWriter, r *http.Request) {
		prepare(config, dataMap, r).ServeHTTP(w, r)
	})

	// TODO maybe we don't need this
	router.Get("/cancel", func(w http.ResponseWriter, r *http.Request) {
		actionID := r.URL.Query().Get("action_id")
		value, loaded := dataMap.Load(actionID)
		if !loaded {
			http.Error(w, "Error action with this is is not known", http.StatusInternalServerError)
			return
		}
		action, ok := value.(Action)
		if !ok {
			http.Error(w, "Error Sidecar in invalid state", http.StatusInternalServerError)
			return
		}
		defer action.cancel()
		resp, err := cancelRequest(config.UserServiceURL, actionID)
		if err != nil {
			http.Error(w, "Error Sidecar in invalid state", http.StatusInternalServerError)
			return
		}
		if resp.StatusCode != 200 {
			http.Error(w, "Error forwarding request or non-200 status received", http.StatusInternalServerError)
		}
	})

	return router
}

func cancelRequest(userServiceURL string, actionID string) (*http.Response, error) {
	req, err := http.NewRequest("DELETE", userServiceURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add(ActionIDHeader, actionID)
	client := &http.Client{}
	return client.Do(req)
}

func setupInternalRouter(dataMap *sync.Map) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Get("/var", func(w http.ResponseWriter, r *http.Request) {
		//TODO:
	})
	router.Get("/log", func(w http.ResponseWriter, r *http.Request) {
		actionID := r.URL.Query().Get(actionIDHeader)
		if actionID == "" {
			http.Error(w, "Missing actionID header", http.StatusBadRequest)
			return
		}
		logLevel := r.URL.Query().Get(LogLevelHeader) // this header is optional
		value, loaded := dataMap.Load(actionID)

		if !loaded {
			http.Error(w, "Error action with this is is not known", http.StatusInternalServerError)
			return
		}
		action, ok := value.(Action)
		if !ok {
			http.Error(w, "Error Sidecar in invalid state", http.StatusInternalServerError)
			return
		}
		actionLog := slog.Debug
		switch logLevel {
		case "ERROR", "error":
			actionLog = slog.Error
		case "WARN", "warn":
			actionLog = slog.Warn
		case "INFO", "info":
			actionLog = slog.Info
		case "DEBUG", "debug":
			actionLog = slog.Debug
		}
		req := make([]byte, 0)
		_, err := r.Body.Read(req)
		if err != nil {
			http.Error(w, "Failed to read the body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()
		actionLog(string(req), "trace", action.Trace, "span", action.Span, "branch", action.Branch, "instance", action.Instance, "namespace", action.Namespace, "state", action.State, "track", "instance."+action.Callpath)
	})
	router.Post("/log", func(w http.ResponseWriter, r *http.Request) {
		actionID := r.URL.Query().Get(actionIDHeader)
		if actionID == "" {
			http.Error(w, "Missing actionID header", http.StatusBadRequest)
			return
		}
		logLevel := r.URL.Query().Get(LogLevelHeader) // this header is optional
		value, loaded := dataMap.Load(actionID)

		if !loaded {
			http.Error(w, "Error action with this is is not known", http.StatusInternalServerError)
			return
		}
		action, ok := value.(Action)
		if !ok {
			http.Error(w, "Error Sidecar in invalid state", http.StatusInternalServerError)
			return
		}
		actionLog := slog.Debug
		switch logLevel {
		case "ERROR", "error":
			actionLog = slog.Error
		case "WARN", "warn":
			actionLog = slog.Warn
		case "INFO", "info":
			actionLog = slog.Info
		case "DEBUG", "debug":
			actionLog = slog.Debug
		}
		req := make([]byte, 0)
		_, err := r.Body.Read(req)
		if err != nil {
			http.Error(w, "Failed to read the body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()
		actionLog(string(req), "trace", action.Trace, "span", action.Span, "branch", action.Branch, "instance", action.Instance, "namespace", action.Namespace, "state", action.State, "track", "instance."+action.Callpath)
	})
	return router
}

type Action struct {
	RequestCarrier
	cancel func()
}

// TODO move to our logic file.
type RequestCarrier struct { // TODO: move this to flow ?
	Deadline  time.Duration `json:"deadline"`
	UserInput []byte        `json:"userInput"`
	Meta
	Data
}

type Meta struct { // TODO: move this to flow ?
	Trace     string `json:"trace"`
	Span      string `json:"span"`
	State     string `json:"state"`
	Branch    string `json:"branch"`
	Instance  string `json:"instance"`
	Workflow  string `json:"workflow"`
	Namespace string `json:"namespace"`
	Callpath  string `json:"callpath"`
}

type Data struct { // TODO: move this to flow ?
	Files FunctionFileDefinition `json:"files"`
}

type FunctionFileDefinition struct { // TODO: import from model pkg instead
	Key         string `json:"key"`
	As          string `json:"as,omitempty"`
	Scope       string `json:"scope,omitempty"`
	Type        string `json:"type,omitempty"`
	Permissions string `json:"permissions,omitempty"`
	Content     string `json:"content,omitempty"`
}

type ResponseCarrier struct {
	UserOutput []byte `json:"userOutput"`
	Err        any    `json:"err"`
	ErrCode    string `json:"errCode"`
}

func prepare(config Config, dataMap *sync.Map, r *http.Request) http.HandlerFunc {

	actionID := r.URL.Query().Get("action_id")
	dec := json.NewDecoder(r.Body)
	var c RequestCarrier
	err := dec.Decode(&c)
	if err != nil {
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}
	}
	ctx, cancel := context.WithCancel(r.Context())
	action := Action{
		RequestCarrier: c,
		cancel:         cancel,
	}
	dataMap.Store(actionID, action)
	filesLocation := filepath.Join(SharedDir, actionID)
	r.Header.Add(FilesLocationHeader, filesLocation)
	return func(w http.ResponseWriter, r *http.Request) {
		buffer := new(bytes.Buffer)
		defer func() {
			dataMap.Delete(actionID)
			os.RemoveAll(filesLocation)
		}()
		_, err := buffer.Write(c.UserInput)
		if err != nil {
			http.Error(w, "Error preparing request body", http.StatusInternalServerError)
			return
		}

		req, err := http.NewRequest("POST", config.UserServiceURL+"?action_id="+actionID, buffer)
		if err != nil {
			http.Error(w, "Error creating new request", http.StatusInternalServerError)
			return
		}

		ctx, cancel2 := context.WithTimeout(ctx, action.Deadline)
		defer cancel2()
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json") // TODO verfify the necessary header.
		select {
		case <-ctx.Done():
			rC := ResponseCarrier{
				ErrCode: "context canceled",
				Err:     ctx.Err(),
			}
			writeJSON(w, rC)
		default:
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				rC := ResponseCarrier{
					ErrCode: "non error response",
					Err:     err,
				}
				writeJSON(w, rC)
				return
			}

			errCode := resp.Header.Get(ErrorCodeHeader)
			errMsg := resp.Header.Get(ErrorMessageHeader)

			if errCode != "" {
				rC := ResponseCarrier{
					ErrCode: errCode,
					Err:     fmt.Errorf(errMsg),
				}
				writeJSON(w, rC)
				return
			}
			capValue := config.MaxResponseSize
			cap, err := strconv.Atoi(capValue)
			if err != nil {
				rC := ResponseCarrier{
					ErrCode: "configuration invalid",
					Err:     err,
				}
				writeJSON(w, rC)
				return

			}
			if resp.ContentLength > int64(cap) {
				rC := ResponseCarrier{
					ErrCode: "content too large",
					Err:     fmt.Errorf("response content is too large"),
				}
				writeJSON(w, rC)
				return
			}

			if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
				rC := ResponseCarrier{
					ErrCode: "container a returned failure status",
					Err:     fmt.Errorf("container failed with status %v", resp.StatusCode),
				}
				writeJSON(w, rC)
				return
			}

			// Forward the response from the managed service back to the original requester.
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				rC := ResponseCarrier{
					ErrCode: "Error reading response body",
					Err:     fmt.Errorf("Error reading response body"),
				}
				writeJSON(w, rC)
				return
			}

			rC := ResponseCarrier{
				UserOutput: body,
			}
			writeJSON(w, rC)
		}

	}
}

func writeFiles(location string, files []FunctionFileDefinition) error {
	// Create the target directory with appropriate permissions
	if err := os.MkdirAll(location, 0o750); err != nil {
		return err
	}

	// Process each file definition
	for _, f := range files {
		path := filepath.Join(location, f.Key)
		data, err := base64.StdEncoding.DecodeString(f.Content)
		if err != nil {
			return err
		}
		// Handle different file types
		switch f.Type {
		case "plain":
			// Write plain text content
			err := os.WriteFile(path, data, 0640)
			if err != nil {
				return err
			}
		case "base64":
			err := os.WriteFile(path, data, 0640)
			if err != nil {
				return err
			}
		case "tar":
			// Extract the contents of a TAR archive
			buf := bytes.NewBuffer(data)
			err := untar(location, f.Permissions, buf)
			if err != nil {
				return err
			}
		case "tar.gz":
			// Call wrapper function for gzip decompression and untar
			err := decompressAndUntar(location, f.Permissions, data)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("unsupported file type: %s", f.Type)
		}
		if f.Permissions != "" {
			p, err := strconv.ParseUint(f.Permissions, 8, 32)
			if err != nil {
				return fmt.Errorf("failed to parse file permissions: %w", err)
			}

			err = os.Chmod(path, os.FileMode(uint32(p)))
			if err != nil {
				return fmt.Errorf("failed to apply file permissions: %w", err)
			}
		}
	}

	return nil
}

func writeOk(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	payLoad := struct {
		Data any `json:"data"`
	}{
		Data: v,
	}
	_ = json.NewEncoder(w).Encode(payLoad)
}

func decompressAndUntar(location string, perms string, encodedData []byte) error {
	gr, err := gzip.NewReader(bytes.NewBuffer(encodedData))
	if err != nil {
		return err
	}
	defer gr.Close()

	// Untar directly from the gzip reader
	return untar(location, perms, gr)
}
