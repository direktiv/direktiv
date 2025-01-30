package api

import (
	"net/http"

	"github.com/direktiv/direktiv/pkg/metastore"
	"github.com/go-chi/chi/v5"
)

type newMetricsCtr struct {
	meta metastore.MetricsStore
}

func (c *newMetricsCtr) mountRouter(r chi.Router) {
	r.Get("/", c.getAll)
	r.Get("/{name}", c.get)
}

func (c *newMetricsCtr) getAll(w http.ResponseWriter, r *http.Request) {
	names, err := c.meta.GetAll(r.Context(), 10000)
	if err != nil {
		writeError(w, &Error{
			Code:    "error",
			Message: err.Error(),
		})

		return
	}
	writeJSON(w, names)
}

func (c *newMetricsCtr) get(w http.ResponseWriter, r *http.Request) {
	names := chi.URLParam(r, "name")
	metrics, err := c.meta.Get(r.Context(), names, metastore.MetricsQueryOptions{
		Limit: 10000,
	})
	if err != nil {
		writeError(w, &Error{
			Code:    "error",
			Message: err.Error(),
		})

		return
	}
	writeJSON(w, metrics)
}
