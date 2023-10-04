package api

import (
	"context"
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/go-chi/chi/v5"
)

type ctxKeyNamespace struct{}

type appMiddlewares struct {
	dStore datastore.Store
}

func (a *appMiddlewares) injectNamespace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		namespace := chi.URLParam(r, "namespace")

		ns, err := a.dStore.Namespaces().GetByName(r.Context(), namespace)
		if err != nil {
			writeDataStoreError(w, err)

			return
		}

		ctx := context.WithValue(r.Context(), ctxKeyNamespace{}, ns)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
