package api

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/go-chi/chi/v5"
)

type appMiddlewares struct {
	db    *database.DB
	cache core.Cache
}

const cacheKey = "api-namespaces"

func (a *appMiddlewares) checkNamespace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		namespace := chi.URLParam(r, "namespace")

		namespaces, err := a.fetchNamespacesFromCache()
		if err != nil {
			writeInternalError(w, err)
			return
		}

		if namespaces == nil {
			namespaces, err = a.fetchNamespacesFromDB(r.Context())
			if err != nil {
				writeInternalError(w, err)
				return
			}
		}

		// if it is not in the list we fetch again
		if !slices.Contains(namespaces, namespace) {
			namespaces, err = a.fetchNamespacesFromDB(r.Context())
			if err != nil {
				writeInternalError(w, err)
				return
			}

			if !slices.Contains(namespaces, namespace) {
				writeError(w, &Error{
					Code:    "resource_not_found",
					Message: "requested resource(namespace) is not found",
				})

				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (a *appMiddlewares) fetchNamespacesFromCache() ([]string, error) {
	ns, exists := a.cache.Get(cacheKey)

	var (
		namespaces []string
		ok         bool
	)

	if exists {
		namespaces, ok = ns.([]string)
		if !ok {
			return namespaces, fmt.Errorf("namespace cache cast error")
		}
	}

	return namespaces, nil
}

func (a *appMiddlewares) fetchNamespacesFromDB(ctx context.Context) ([]string, error) {
	nsList, err := a.db.DataStore().Namespaces().GetAll(ctx)
	if err != nil {
		return []string{}, err
	}
	names := make([]string, 0)
	for i := range nsList {
		names = append(names, nsList[i].Name)
	}
	a.cache.Set(cacheKey, names)

	return names, nil
}
