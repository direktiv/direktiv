package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/direktiv/direktiv/pkg/sidecar2/types"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func setupAPIforUserContainer(dataMap *sync.Map) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Get("/var", func(w http.ResponseWriter, r *http.Request) {
		// TODO: .
	})
	router.Get("/log", func(w http.ResponseWriter, r *http.Request) {
		// status, err := handleGetLog(r, dataMap)
		// if err != nil {
		// 	http.Error(w, err.Error(), status)
		// }
	})
	router.Post("/log", func(w http.ResponseWriter, r *http.Request) {
		status, err := handlePostLog(r, dataMap)
		if err != nil {
			http.Error(w, err.Error(), status)
		}
	})

	return router
}

func handlePostLog(r *http.Request, dataMap *sync.Map) (int, error) {
	actionID := r.URL.Query().Get(ActionIDHeader)
	if actionID == "" {
		return http.StatusBadRequest, fmt.Errorf("missing actionID header")
	}
	logLevel := r.URL.Query().Get(LogLevelHeader)
	value, loaded := dataMap.Load(actionID)

	if !loaded {
		return http.StatusInternalServerError, fmt.Errorf("failed to handle request action with this ID is not known")
	}
	action, ok := value.(types.ActionController)
	if !ok {
		return http.StatusInternalServerError, fmt.Errorf("failed to handle request sidecar in invalid state")
	}
	req := make([]byte, 0)
	_, err := r.Body.Read(req)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to read the body")
	}
	defer r.Body.Close()
	msg := string(req)
	logMsg(logLevel, msg, action)

	return http.StatusOK, nil
}

func logMsg(logLevel string, msg string, action types.ActionController) {
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
	actionLog(msg, "trace", action.Trace,
		"span", action.Span, "branch",
		action.Branch, "instance", action.Instance,
		"namespace", action.Namespace,
		"state", action.State,
		"track", "instance."+action.Callpath)
}
