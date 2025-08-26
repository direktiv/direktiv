package api

import (
	"context"
	"net/http"
	"slices"

	"github.com/direktiv/direktiv/pkg/cache"
	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/go-chi/chi/v5"
)

type appMiddlewares struct {
	dStore datastore.Store
	cache  *cache.Cache
}

const cacheKey = "api-namespaces"

func (a *appMiddlewares) checkNamespace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		namespace := chi.URLParam(r, "namespace")

		list := a.fetchNamespacesFromCache()
		var err error

		if list == nil || !slices.Contains(list, namespace) {
			list, err = a.fetchNamespacesFromDB(r.Context())
			if err != nil {
				writeInternalError(w, err)
				return
			}
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

func (a *appMiddlewares) fetchNamespacesFromCache() []string {
	nsList, exists := a.cache.Get(cacheKey)
	if exists {
		return nsList.([]string)
	}

	return nil
}

func (a *appMiddlewares) fetchNamespacesFromDB(ctx context.Context) ([]string, error) {
	nsList, err := a.dStore.Namespaces().GetAll(ctx)
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
