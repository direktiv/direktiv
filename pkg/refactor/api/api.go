package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/middlewares"
	"github.com/go-chi/chi/v5"
)

const (
	shutdownWaitTime  = 5 * time.Second
	readHeaderTimeout = 5 * time.Second
)

func Start(app core.App, db *database.DB, addr string, done <-chan struct{}, wg *sync.WaitGroup) {
	funcCtr := &serviceController{
		manager: app.ServiceManager,
	}

	regCtr := &registryController{
		manager: app.RegistryManager,
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

	r.Handle("/api/v2/gw/*", app.EndpointManager)

	r.Get("/api/v2/version", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, app.Version)
	})

	r.Route("/api/v2", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			var chiMiddlewares []func(http.Handler) http.Handler
			for i := range middlewares.GetMiddlewares() {
				chiMiddlewares = append(chiMiddlewares, middlewares.GetMiddlewares()[i])
			}
			chiMiddlewares = append(chiMiddlewares, mw.injectNamespace)

			r.Use(chiMiddlewares...)

			r.Route("/namespaces/{namespace}/services", func(r chi.Router) {
				funcCtr.mountRouter(r)
			})

			r.Route("/namespaces/{namespace}/registries", func(r chi.Router) {
				regCtr.mountRouter(r)
			})

			r.Get("/namespaces/{namespace}/endpoints", func(w http.ResponseWriter, r *http.Request) {
				writeJSON(w, app.EndpointManager.GetAll())
			})

			r.Get("/namespaces/{namespace}/plugins", func(w http.ResponseWriter, r *http.Request) {
				data, err := app.GetAllPluginSchemas()
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
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	go func() {
		// Run api server
		err := apiServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
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
			log.Fatal(err)
		}
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
	// nolint:errchkjson
	_ = json.NewEncoder(w).Encode(payLoad)
}

func writeOk(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
