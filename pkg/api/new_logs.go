package api

import (
	"net/http"
	"time"

	"github.com/direktiv/direktiv/pkg/metastore"
	"github.com/go-chi/chi/v5"
)

type newLogsCtr struct {
	meta metastore.LogStore
}

func (c *newLogsCtr) mountRouter(r chi.Router) {
	r.Get("/", c.get)
}

func (c *newLogsCtr) get(w http.ResponseWriter, r *http.Request) {
	logs, err := c.meta.Get(r.Context(), metastore.LogQueryOptions{
		StartTime: time.Now().Add(-time.Hour),
		EndTime:   time.Now(),
	})
	if err != nil {
		writeError(w, &Error{
			Code:    "error",
			Message: err.Error(),
		})

		return
	}
	writeJSON(w, logs)
}
