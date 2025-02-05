package api

import (
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/go-chi/chi/v5"
)

const ctxKeyNamespace = "namespace"

type appMiddlewares struct {
	dStore datastore.Store
}

func (a *appMiddlewares) injectNamespace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		namespace := chi.URLParam(r, "namespace")

		ns, err := a.dStore.Namespaces().GetByName(r.Context(), namespace)
		if errors.Is(err, datastore.ErrNotFound) {
			writeError(w, &Error{
				Code:    "resource_not_found",
				Message: "requested resource(namespace) is not found",
			})

			return
		}
		if err != nil {
			writeInternalError(w, err)

			return
		}

		ctx := context.WithValue(r.Context(), ctxKeyNamespace, ns)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *appMiddlewares) checkAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//nolint:gosec
		apiKeyHeader := "Direktiv-Api-Key"
		apiKey := os.Getenv("DIREKTIV_API_KEY")
		if apiKey != "" && apiKey != r.Header.Get(apiKeyHeader) {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		next.ServeHTTP(w, r)
	})
}
