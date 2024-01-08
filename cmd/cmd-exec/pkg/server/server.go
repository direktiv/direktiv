package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
)

const (
	DirektivActionIDHeader = "Direktiv-ActionID"
	DirektivTempDir        = "Direktiv-TempDir"

	DirektivErrorCodeHeader    = "Direktiv-ErrorCode"
	DirektivErrorMessageHeader = "Direktiv-ErrorMessage"

	DirektivErrorCode = "io.direktiv.error.execution"
)

type Server[IN any] struct {
	httpServer *http.Server
	stopChan   chan os.Signal
}

type ExecutionInfo struct {
	TmpDir string
	Log    *Logger
}

func NewServer[IN any](fn func(context.Context, IN, *ExecutionInfo) (interface{}, error)) *Server[IN] {
	server := &http.Server{
		Addr:         "0.0.0.0:8080",
		Handler:      Handler[IN](fn),
		ReadTimeout:  1 * time.Minute,
		WriteTimeout: 4 * time.Hour,
		IdleTimeout:  15 * time.Second,
	}

	return &Server[IN]{
		httpServer: server,
		stopChan:   make(chan os.Signal, 2),
	}
}

func Handler[IN any](fn func(context.Context, IN, *ExecutionInfo) (interface{}, error)) http.Handler {
	r := chi.NewRouter()

	errWriter := func(w http.ResponseWriter, status int, errMsg string) {
		w.WriteHeader(status)
		w.Header().Set(DirektivErrorCodeHeader, DirektivErrorCode)
		w.Header().Set(DirektivErrorMessageHeader, errMsg)

		w.Write([]byte(errMsg))
	}

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {

		var data IN

		b, err := io.ReadAll(r.Body)
		if err != nil {
			errWriter(w, http.StatusBadRequest, err.Error())

			return
		}
		defer r.Body.Close()

		if len(b) > 0 {
			err = json.Unmarshal(b, &data)
			if err != nil {
				errWriter(w, http.StatusBadRequest, err.Error())

				return
			}
			defer r.Body.Close()
		}

		// get tmp dir
		tmpDir := r.Header.Get(DirektivTempDir)
		if tmpDir == "" {
			errWriter(w, http.StatusBadRequest, "no temp directory provided")

			return

		}

		actionID := r.Header.Get(DirektivActionIDHeader)
		if actionID == "" {
			errWriter(w, http.StatusBadRequest, "no action id provided")

			return
		}

		ei := &ExecutionInfo{
			TmpDir: tmpDir,
			Log:    NewLogger(actionID),
		}

		out, err := fn(r.Context(), data, ei)
		if err != nil {
			errWriter(w, http.StatusInternalServerError, err.Error())

			return
		}

		w.Header().Add("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(out); err != nil {
			errWriter(w, http.StatusInternalServerError, err.Error())

			return
		}

	})

	return r
}

func (s *Server[IN]) Start() {

	slog.Info("starting server")

	signal.Notify(s.stopChan,
		syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("error starting server", slog.String("error", err.Error()))

			os.Exit(1)
		}
	}()

	<-s.stopChan
	s.Stop()
}

func (s *Server[IN]) Stop() {

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
