// nolint
package webapi

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/function2"
	"github.com/go-chi/chi/v5"
)

func Start(funcManager *function2.Manager, addr string, done <-chan struct{}, wg *sync.WaitGroup) {
	fcnt := &functionsController{
		manager: funcManager,
	}
	r := chi.NewRouter()

	r.Route("/api/v2", func(r chi.Router) {
		r.Route("/functions", func(r chi.Router) {
			fcnt.mountRouter(r)
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

func writeData(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	payLoad := struct {
		Data any
	}{
		Data: v,
	}
	_ = json.NewEncoder(w).Encode(payLoad)
}
