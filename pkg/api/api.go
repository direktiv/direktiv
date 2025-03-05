package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/events"
	"github.com/direktiv/direktiv/pkg/extensions"
	"github.com/direktiv/direktiv/pkg/instancestore"
	pubsub2 "github.com/direktiv/direktiv/pkg/pubsub"
	"github.com/direktiv/direktiv/pkg/tracing"
	"github.com/direktiv/direktiv/pkg/version"
	"github.com/go-chi/chi/v5"
)

const (
	shutdownWaitTime  = 5 * time.Second
	readHeaderTimeout = 5 * time.Second
)

func Initialize(circuit *core.Circuit, app core.App, db *database.DB, bus *pubsub2.Bus, instanceManager *instancestore.InstanceManager, wakeByEvents events.WakeEventsWaiter, startByEvents events.WorkflowStart) error {
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
		db:              db,
		bus:             bus,
		registryManager: app.RegistryManager,
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

	jxCtr := jxController{}

	mw := &appMiddlewares{dStore: db.DataStore()}

	r := chi.NewRouter()
	r.Use(tracing.OtelMiddleware())
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

	for _, extraRoute := range GetExtraRoutes() {
		extraRoute(r)
	}

	// version endpoint
	r.Get("/api/v2/status", func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Version      string `json:"version"`
			IsEnterprise bool   `json:"isEnterprise"`
			RequiresAuth bool   `json:"requiresAuth"`
		}{
			Version:      version.Version,
			IsEnterprise: extensions.IsEnterprise,
			RequiresAuth: os.Getenv("DIREKTIV_UI_SET_API_KEY") == "true",
		}

		writeJSON(w, data)
	})

	logCtr := &logController{
		store:       db.DataStore().NewLogs(),
		logsBackend: app.Config.LogsBackend,
	}
	r.Handle("/ns/{namespace}/*", app.GatewayManager)

	r.Route("/api/v2", func(r chi.Router) {
		r.Route("/namespaces", func(r chi.Router) {
			if extensions.CheckOidcMiddleware != nil {
				r.Use(
					extensions.CheckOidcMiddleware,
					extensions.CheckAPITokenMiddleware,
					extensions.CheckAPIKeyMiddleware)
			}
			nsCtr.mountRouter(r)
		})

		r.Group(func(r chi.Router) {
			if extensions.CheckOidcMiddleware != nil {
				r.Use(
					extensions.CheckOidcMiddleware,
					extensions.CheckAPITokenMiddleware,
					extensions.CheckAPIKeyMiddleware)
			}
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
				logCtr.mountRouter(r, app.Config)
			})
			r.Route("/namespaces/{namespace}/notifications", func(r chi.Router) {
				notificationsCtr.mountRouter(r)
			})
			r.Route("/namespaces/{namespace}/metrics", func(r chi.Router) {
				metricsCtr.mountRouter(r)
			})
			r.Route("/namespaces/{namespace}/events/history", func(r chi.Router) {
				eventsCtr.mountEventHistoryRouter(r)
			})
			r.Route("/namespaces/{namespace}/events/listeners", func(r chi.Router) {
				eventsCtr.mountEventListenerRouter(r)
			})
			r.Route("/namespaces/{namespace}/events/broadcast", func(r chi.Router) {
				eventsCtr.mountBroadcast(r)
			})
			r.Handle("/namespaces/{namespace}/gateway/*", app.GatewayManager)

			if len(extensions.AdditionalAPIRoutes) > 0 {
				for pattern, ctr := range extensions.AdditionalAPIRoutes {
					r.Route(pattern, func(r chi.Router) {
						ctr(r)
					})
				}
			}
		})

		r.Route("/jx", func(r chi.Router) {
			jxCtr.mountRouter(r)
		})
	})

	r.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/api/") {
			writeError(w, &Error{
				Code:    "request_path_not_found",
				Message: "request http path is not found",
			})

			return
		}

		staticDir := "/app/ui"
		fs := http.FileServer(http.Dir(staticDir))

		if _, err := os.Stat(staticDir + r.URL.Path); err == nil {
			fs.ServeHTTP(w, r)
			return
		}

		http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
	})

	apiServer := &http.Server{
		Addr:              "0.0.0.0:" + strconv.Itoa(app.Config.ApiPort),
		Handler:           r,
		ReadHeaderTimeout: readHeaderTimeout,
	}

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
	ns := r.Context().Value(ctxKeyNamespace).(*datastore.Namespace)

	return ns
}
