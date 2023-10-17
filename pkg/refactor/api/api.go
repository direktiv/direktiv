// nolint
package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/go-chi/chi/v5"
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
		return
	})
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		writeError(w, &Error{
			Code:    "request_path_not_found",
			Message: "request http path is not found",
		})

		return
	})
	r.Get("/api/v2/version", func(w http.ResponseWriter, r *http.Request) {
		writeJson(w, app.Version)
	})
	r.Route("/api/v2", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(mw.injectNamespace)

			r.Route("/namespaces/{namespace}/services", func(r chi.Router) {
				funcCtr.mountRouter(r)
			})

			r.Route("/namespaces/{namespace}/registries", func(r chi.Router) {
				regCtr.mountRouter(r)
			})
		})
	})

	apiServer := &http.Server{Addr: addr, Handler: r}
	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	go func() {
		// Run api server
		err := apiServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
		// Wait for server context to be stopped
		<-serverCtx.Done()
		wg.Done()
	}()

	go func() {
		<-done
		// Shutdown signal with grace period of 5 seconds
		shutdownCtx, _ := context.WithTimeout(serverCtx, 5*time.Second)
		err := apiServer.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()
}

func writeJson(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	payLoad := struct {
		Data any `json:"data"`
	}{
		Data: v,
	}
	_ = json.NewEncoder(w).Encode(payLoad)
}

func writeOk(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
