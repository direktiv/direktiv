package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/direktiv/direktiv/pkg/refactor/events"
	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
	"github.com/direktiv/direktiv/pkg/refactor/middlewares"
	pubsub2 "github.com/direktiv/direktiv/pkg/refactor/pubsub"
	"github.com/direktiv/direktiv/pkg/version"
	"github.com/go-chi/chi/v5"
)

const (
	shutdownWaitTime  = 5 * time.Second
	readHeaderTimeout = 5 * time.Second
)

func Initialize(app core.App, db *database.SQLStore, bus *pubsub2.Bus, instanceManager *instancestore.InstanceManager, wakeByEvents events.WakeEventsWaiter, startByEvents events.WorkflowStart, addr string, circuit *core.Circuit) error {
	funcCtr := &serviceController{
		manager: app.ServiceManager,
	}
	fsCtr := &fsController{
		db:  db,
		bus: bus,
	}
	regCtr := &registryController{
		manager: app.RegistryManager,
	}
	varCtr := &varController{
		db: db,
	}
	secCtr := &secretsController{
		db: db,
	}
	nsCtr := &nsController{
		db:  db,
		bus: bus,
	}
	mirrorsCtr := &mirrorsController{
		db:            db,
		bus:           bus,
		syncNamespace: app.SyncNamespace,
	}
	instCtr := &instController{
		db:      db,
		manager: instanceManager,
	}
	notificationsCtr := &notificationsController{
		db: db,
	}
	metricsCtr := &metricsController{
		db: db,
	}
	eventsCtr := eventsController{
		store:         db.DataStore(),
		wakeInstance:  wakeByEvents,
		startWorkflow: startByEvents,
	}

	mw := &appMiddlewares{dStore: db.DataStore()}

	r := chi.NewRouter()
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		writeError(w, &Error{
			Code:    "request_method_not_allowed",
			Message: "request http method is not allowed for this path",
		})
	})
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		writeError(w, &Error{
			Code:    "request_path_not_found",
			Message: "request http path is not found",
		})
	})

	chiMiddlewares := make([]func(http.Handler) http.Handler, 0)
	for i := range middlewares.GetMiddlewares() {
		chiMiddlewares = append(chiMiddlewares, middlewares.GetMiddlewares()[i])
	}

	r.Use(chiMiddlewares...)

	for _, extraRoute := range GetExtraRoutes() {
		extraRoute(r)
	}

	// handle namespace and gateway
	r.Handle("/ns/{namespace}/*", app.GatewayManager)

	// version endpoint
	r.Get("/api/v2/status", func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Version      string `json:"version"`
			IsEnterprise bool   `json:"isEnterprise"`
			RequiresAuth bool   `json:"requiresAuth"`
		}{
			Version:      version.Version,
			IsEnterprise: app.Config.IsEnterprise,
			RequiresAuth: os.Getenv("DIREKTIV_API_KEY") != "",
		}

		writeJSON(w, data)
	})

	logCtr := &logController{
		store: db.DataStore().NewLogs(),
	}

	r.Route("/api/v2", func(r chi.Router) {
		r.Route("/namespaces", func(r chi.Router) {
			nsCtr.mountRouter(r)
		})

		r.Group(func(r chi.Router) {
			r.Use(mw.injectNamespace)

			r.Route("/namespaces/{namespace}/instances", func(r chi.Router) {
				instCtr.mountRouter(r)
			})
			r.Route("/namespaces/{namespace}/syncs", func(r chi.Router) {
				mirrorsCtr.mountRouter(r)
			})
			r.Route("/namespaces/{namespace}/secrets", func(r chi.Router) {
				secCtr.mountRouter(r)
			})
			r.Route("/namespaces/{namespace}/variables", func(r chi.Router) {
				varCtr.mountRouter(r)
			})
			r.Route("/namespaces/{namespace}/files", func(r chi.Router) {
				fsCtr.mountRouter(r)
			})
			r.Route("/namespaces/{namespace}/services", func(r chi.Router) {
				funcCtr.mountRouter(r)
			})
			r.Route("/namespaces/{namespace}/registries", func(r chi.Router) {
				regCtr.mountRouter(r)
			})
			r.Route("/namespaces/{namespace}/logs", func(r chi.Router) {
				logCtr.mountRouter(r)
			})
			r.Route("/namespaces/{namespace}/notifications", func(r chi.Router) {
				notificationsCtr.mountRouter(r)
			})
			r.Route("/namespaces/{namespace}/metrics", func(r chi.Router) {
				metricsCtr.mountRouter(r)
			})
			r.Get("/namespaces/{namespace}/gateway/consumers", func(w http.ResponseWriter, r *http.Request) {
				data, err := app.GatewayManager.GetConsumers(chi.URLParam(r, "namespace"))
				if err != nil {
					writeInternalError(w, err)

					return
				}
				writeJSON(w, data)
			})
			r.Get("/namespaces/{namespace}/gateway/routes", func(w http.ResponseWriter, r *http.Request) {
				data, err := app.GatewayManager.GetRoutes(chi.URLParam(r, "namespace"), r.URL.Query().Get("path"))
				if err != nil {
					writeInternalError(w, err)

					return
				}
				writeJSON(w, data)
			})
			r.Route("/namespaces/{namespace}/events/history", func(r chi.Router) {
				eventsCtr.mountEventHistoryRouter(r)
			})
			r.Route("/namespaces/{namespace}/events/listener", func(r chi.Router) {
				eventsCtr.mountEventListenerRouter(r)
			})
			r.Route("/namespaces/{namespace}/events/broadcast", func(r chi.Router) {
				eventsCtr.mountBroadcast(r)
			})
		})
	})

	apiServer := &http.Server{Addr: addr, Handler: r, ReadHeaderTimeout: readHeaderTimeout}

	circuit.Start(func() error {
		err := apiServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listen on port, err: %w", err)
		}

		return nil
	})

	circuit.OnCancel(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownWaitTime)
		err := apiServer.Shutdown(shutdownCtx)
		if err != nil {
			slog.Error("api v2 shutdown server", "err", err)
			panic(err)
		}
		cancel()
	})

	// TODO: setup shutdown handler.

	return nil
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

func writeJSONWithMeta(w http.ResponseWriter, data any, meta any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	payload := struct {
		Meta any `json:"meta"`
		Data any `json:"data"`
	}{
		Data: data,
		Meta: meta,
	}
	_ = json.NewEncoder(w).Encode(payload)
}

func writeOk(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func extractContextNamespace(r *http.Request) *datastore.Namespace {
	//nolint:forcetypeassert
	ns := r.Context().Value(ctxKeyNamespace{}).(*datastore.Namespace)

	return ns
}
