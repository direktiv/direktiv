package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/internal/telemetry"
	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
)

const (
	DirektivErrorCode = "io.direktiv.error.execution"
)

type File struct {
	Name       string `json:"name"`
	Content    string `json:"content"`
	Permission uint   `json:"permission"`
}

type Payload struct {
	Files    []File    `json:"files"`
	Commands []Command `json:"commands"`
}

type Server struct {
	httpServer *http.Server
	stopChan   chan os.Signal
}

type ExecutionInfo struct {
	TmpDir string
	Log    *Logger
}

func NewServer() *Server {
	server := &http.Server{
		Addr:         "0.0.0.0:8080",
		Handler:      Handler(),
		ReadTimeout:  1 * time.Minute,
		WriteTimeout: 4 * time.Hour,
		IdleTimeout:  15 * time.Second,
	}

	return &Server{
		httpServer: server,
		stopChan:   make(chan os.Signal, 2),
	}
}

func errWriter(w http.ResponseWriter, status int, errMsg string) {
	slog.Error("writing error response", slog.Int("status", status), slog.String("error", errMsg))
	w.Header().Set(core.EngineHeaderErrorCode, DirektivErrorCode)
	w.Header().Set(core.EngineHeaderErrorMessage, errMsg)

	w.WriteHeader(status)

	_, err := w.Write([]byte(errMsg))
	if err != nil {
		slog.Error("failed to write error response", slog.String("error", err.Error()))
	}
}

func Handler() http.Handler {
	r := chi.NewRouter()

	r.Get("/healthz", readinessHandler)
	r.Get("/readiness", readinessHandler)

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		var in Payload
		b, err := io.ReadAll(r.Body)
		if err != nil {
			errWriter(w, http.StatusBadRequest, "failed to read request body")
			return
		}
		defer r.Body.Close()

		if len(b) > 0 {
			if err := json.Unmarshal(b, &in); err != nil {
				errWriter(w, http.StatusBadRequest, "failed to unmarshal request payload")
				return
			}
		}

		tmpDir := r.Header.Get(core.EngineHeaderTempDir)
		if tmpDir == "" {
			errWriter(w, http.StatusBadRequest, "no temp directory provided")
			return
		}

		actionID := r.Header.Get(core.EngineHeaderActionID)
		if actionID == "" {
			errWriter(w, http.StatusBadRequest, "no action id provided")
			return
		}

		ctx := otel.GetTextMapPropagator().Extract(
			r.Context(),
			propagation.HeaderCarrier(r.Header),
		)

		tracer := otel.Tracer("action-call")
		ctx, span := tracer.Start(ctx, "action-call")
		span.SetAttributes(attribute.KeyValue{
			Key:   "instance",
			Value: attribute.StringValue(actionID),
		},
		)
		defer span.End()

		// setup logging
		lo := telemetry.LogObjectFromHeader(ctx, r.Header)
		ctx = telemetry.LogInitCtx(ctx, lo)
		telemetry.LogInstance(ctx, telemetry.LogLevelInfo, "cmd container executing")

		// backend := "http://localhost:8889"
		// if envBackend := os.Getenv("httpBackend"); envBackend != "" {
		// 	backend = envBackend
		// }

		ei := &ExecutionInfo{
			TmpDir: tmpDir,
			Log:    NewLogger(lo, actionID),
		}

		for _, file := range in.Files {
			err := prepareFile(filepath.Join(tmpDir, file.Name), file.Content, file.Permission)
			if err != nil {
				errWriter(w, http.StatusInternalServerError, "failed to prepare file: "+err.Error())
				return
			}
		}

		out, err := RunCommands(ctx, in, ei)
		if err != nil {
			errWriter(w, http.StatusInternalServerError, "handler function error: "+err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(out); err != nil {
			errWriter(w, http.StatusInternalServerError, "failed to encode response: "+err.Error())
			return
		}
	})

	return r
}

func prepareFile(path, content string, perm uint) error {
	slog.Debug("preparing file", slog.String("path", path), slog.Any("permission", perm))
	file, err := os.Create(path)
	if err != nil {
		slog.Error("failed to create file", slog.String("path", path), slog.String("error", err.Error()))
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		slog.Error("failed to write file content", slog.String("path", path), slog.String("error", err.Error()))
		return err
	}

	if err := file.Chmod(fs.FileMode(perm)); err != nil {
		slog.Error("failed to set file permissions", slog.String("path", path), slog.String("error", err.Error()))
		return err
	}

	slog.Info("file prepared successfully", slog.String("path", path))

	return nil
}

func (s *Server) Start() {
	slog.Info("starting server")

	signal.Notify(s.stopChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("error starting server", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	<-s.stopChan
	s.Stop()
}

func (s *Server) Stop() {
	slog.Info("stopping server")
	s.httpServer.SetKeepAlivesEnabled(false)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go func() {
		<-shutdownCtx.Done()
		if shutdownCtx.Err() == context.DeadlineExceeded {
			slog.Error("shutdown timed out")
			os.Exit(1)
		}
	}()

	err := s.httpServer.Shutdown(shutdownCtx)
	if err != nil {
		slog.Error("shutdown failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("server stopped")
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
