package api

import (
	"context"
	"net/http"
	"slices"

	"github.com/direktiv/direktiv/internal/cluster/cache"
	"github.com/direktiv/direktiv/internal/datastore/datasql"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type appMiddlewares struct {
	db    *gorm.DB
	cache cache.Manager
}

const cacheKey = "api-namespaces"

func (a *appMiddlewares) checkNamespace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		namespace := chi.URLParam(r, "namespace")

		list, err := a.cache.NamespaceCache().Get("namespaces", func(args ...any) ([]string, error) {
			return a.fetchNamespacesFromDB(r.Context())
		})
		if err != nil {
			writeInternalError(w, err)
			return
		}

		if !slices.Contains(list, namespace) {
			writeError(w, &Error{
				Code:    "resource_not_found",
				Message: "requested resource(namespace) is not found",
			})

			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *appMiddlewares) fetchNamespacesFromDB(ctx context.Context) ([]string, error) {
	nsList, err := datasql.NewStore(a.db).Namespaces().GetAll(ctx)
	if err != nil {
		return []string{}, err
	}
	names := make([]string, 0)
	for i := range nsList {
		names = append(names, nsList[i].Name)
	}

	return names, nil
}
