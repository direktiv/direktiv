package cmdserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mattn/go-shellwords"
)

type Env struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (e Env) toKV() string {
	return fmt.Sprintf("%s=%s", e.Name, e.Value)
}

type Command struct {
	Command  string `json:"command"`
	Continue bool   `json:"continue"`
	Print    bool   `json:"print"`
	Stdout   bool   `json:"stdout"`
	Omit     bool   `json:"omit"`
	Envs     []Env  `json:"envs"`
}

type File struct {
	Name       string `json:"name"`
	Content    string `json:"content"`
	Permission uint   `json:"permission"`
}

type Request struct {
	Commands []Command `json:"commands"`
	Files    []File    `json:"files"`
}

type CommandResponse struct {
	Error  string      `json:"error"`
	Output interface{} `json:"output"`
}

const (
	DirektivActionIDHeader = "Direktiv-ActionID"
	DirektivTempDir        = "Direktiv-TempDir"

	DirektivErrorCodeHeader    = "Direktiv-ErrorCode"
	DirektivErrorMessageHeader = "Direktiv-ErrorMessage"

	DirektivErrorCode = "io.direktiv.error.execution"
)

func RunApplication() {

	// TODO: INIT scripts

	stopChan := make(chan os.Signal, 1)

	portString := os.Getenv("DIREKTIV_PORT")
	if portString == "" {
		panic("no port provided in DIREKTIV_PORT")
	}

	server := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%s", portString),
		Handler:      Handler(),
		ReadTimeout:  1 * time.Minute,
		WriteTimeout: 4 * time.Hour,
		IdleTimeout:  15 * time.Second,
	}

	slog.Info("starting server")

	signal.Notify(stopChan,
		syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("error starting server", slog.String("error", err.Error()))

			os.Exit(1)
		}
	}()

	<-stopChan
	Stop(server)

}

func Handler() http.Handler {

	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
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

		b, err := io.ReadAll(r.Body)
		if err != nil {
			errWriter(w, http.StatusBadRequest, err.Error())

			return
		}
		defer r.Body.Close()

		var req Request
		if len(b) > 0 {
			err = json.Unmarshal(b, &req)
			if err != nil {
				errWriter(w, http.StatusBadRequest, err.Error())

				return
			}
		}

		// handle files
		for a := range req.Files {
			f := req.Files[a]
			err = prepareFile(filepath.Join(tmpDir, f.Name), f.Content, f.Permission)
			if err != nil {
				errWriter(w, http.StatusInternalServerError, err.Error())

				return
			}
		}

		responses := make([]*CommandResponse, 0)
		for a := range req.Commands {
			responses = append(responses, runCommand(req.Commands[a], tmpDir))
		}

		w.Header().Add("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(responses); err != nil {
			errWriter(w, http.StatusInternalServerError, err.Error())

			return
		}
	})

	r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
		// TODO:
	})

	return r
}

func Stop(server *http.Server) {
	slog.Info("stopping server")
	server.SetKeepAlivesEnabled(false)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go func() {
		<-shutdownCtx.Done()
		if shutdownCtx.Err() == context.DeadlineExceeded {
			slog.Error("shutdown timed out")

			os.Exit(1)
		}
	}()

	err := server.Shutdown(shutdownCtx)
	if err != nil {
		slog.Error("shutdown failed", slog.String("error", err.Error()))

		os.Exit(1)
	}

	slog.Info("server stopped")
}

func errWriter(w http.ResponseWriter, status int, errMsg string) {
	w.Header().Set(DirektivErrorCodeHeader, DirektivErrorCode)
	w.Header().Set(DirektivErrorMessageHeader, errMsg)

	w.WriteHeader(status)

	// nolint
	w.Write([]byte(errMsg))
}

func runCommand(cmd Command, dir string) *CommandResponse {

	logger := newLogger("", "")

	response := &CommandResponse{}

	p := shellwords.NewParser()
	p.ParseEnv = true
	p.ParseBacktick = true

	args, err := p.Parse(cmd.Command)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	if len(args) == 0 {
		response.Error = "no binary provided"
	}

	// always a binary
	bin := args[0]
	argsIn := []string{}
	if len(args) > 1 {
		argsIn = args[1:]
	}

	exe := exec.CommandContext(context.Background(), bin, argsIn...)
	exe.Dir = dir
	exe.Stdout = logger
	exe.Stderr = logger

	envs := make([]string, 0)
	envs = append(envs, fmt.Sprintf("HOME=%s", dir))
	envs = append(envs, os.Environ()...)

	for i := range cmd.Envs {
		envs = append(envs, cmd.Envs[i].toKV())
	}
	exe.Env = envs

	err = exe.Run()
	if err != nil {
		response.Error = err.Error()
	}

	response.Output = logger.LogData.String()

	return response
}

func prepareFile(path, content string, perm uint) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	return file.Chmod(fs.FileMode(perm))
}
