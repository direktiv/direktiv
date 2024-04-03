package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
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

func Start(ctx context.Context, app core.App, db *database.DB, bus *pubsub2.Bus, instanceManager *instancestore.InstanceManager, addr string, done <-chan struct{}, wg *sync.WaitGroup) {
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
		db:  db,
		bus: bus,
	}
	instCtr := &instController{
		db:      db,
		manager: instanceManager,
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
	r.Handle("/gw/*", app.GatewayManager)
	r.Handle("/ns/{namespace}/*", app.GatewayManager)

	r.Get("/api/v2/version", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, version.Version)
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
		})
	})

	apiServer := &http.Server{Addr: addr, Handler: r, ReadHeaderTimeout: readHeaderTimeout}
	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(ctx)

	go func() {
		// Run api server
		err := apiServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Debug("API v2 Server Closed", "error", err)
			log.Fatal(err)
		}
		// Wait for server context to be stopped
		<-serverCtx.Done()
		wg.Done()
	}()

	go func() {
		<-done
		shutdownCtx, cancel := context.WithTimeout(serverCtx, shutdownWaitTime)
		defer cancel()

		err := apiServer.Shutdown(shutdownCtx)
		if err != nil {
			slog.Error("Failed to start API server", "addr", addr, "error", err)
			panic(err)
		}
		slog.Debug("Shutting down API server", "addr", addr)
		serverStopCtx()
	}()
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
