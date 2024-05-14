package api

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"sync"

	"github.com/direktiv/direktiv/pkg/flow"
	"github.com/direktiv/direktiv/pkg/refactor/engine"
	"github.com/direktiv/direktiv/pkg/sidecar2/types"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	ActionIDHeader      = "Direktiv-ActionID"
	LogLevelHeader      = "Direktiv-LogLevel"
	FilesLocationHeader = "Direktiv-TempDir"
	ErrorCodeHeader     = "Direktiv-ErrorCode"
	ErrorMessageHeader  = "Direktiv-ErrorMessage"
	ActionIDQuerryParam = "action_id"

	SharedDir = "/mnt/shared"
)

func StartApis(config types.Config, actionCtl *sync.Map) {
	limit, err := strconv.Atoi(config.MaxResponseSize)
	if err != nil {
		slog.Error("parsing config.MaxResponseSize", "error", err, "MaxResponseSize", config.MaxResponseSize)
		panic(err)
	}

	slog.Debug("Initializing sidecar", "MaxResponseSize", limit, "FlowServerURL", config.FlowServerURL)

	slog.Debug("Initializing flow exposed routes")
	externalRouter := setupAPIForFlow(config.UserServiceURL, limit, actionCtl)

	slog.Debug("Initializing user container exposed routes")
	// Internal router, accessible only to the user service.
	internalRouter := setupAPIforUserContainer(actionCtl)

	// Start routers in separate goroutines to listen on different ports.
	go func() {
		log.Fatal(http.ListenAndServe(":"+config.ExternalPort, externalRouter))
		slog.Debug("Started external routes", "port", config.ExternalPort)
	}()

	log.Fatal(http.ListenAndServe("0.0.0.0:"+config.InternalPort, internalRouter))
	slog.Debug("Started internal routes", "addr", "0.0.0.0", "port", config.InternalPort)
}

func setupAPIForFlow(userServiceURL string, maxResponseSize int, actionCtl *sync.Map) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	// Router for handling external requests.
	router.Post("/", func(w http.ResponseWriter, r *http.Request) {
		executeFunction(r, w, userServiceURL, maxResponseSize, actionCtl, extractActionData)
	})

	return router
}

func extractActionData(r *http.Request) (string, engine.ActionRequest, error) {
	aid := r.Header.Get(flow.DirektivActionIDHeader)
	if len(aid) == 0 {
		return "", engine.ActionRequest{}, fmt.Errorf("action id header was not set")
	}
	ar, err := engine.DecodeActionRequest(r)
	if err != nil {
		return "", engine.ActionRequest{}, err
	}

	return aid, ar, nil
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
