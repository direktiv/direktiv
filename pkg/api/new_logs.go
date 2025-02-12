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

func (m *newLogsCtr) mountRouter(r chi.Router) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		logs, err := m.meta.Get(r.Context(), metastore.LogQueryOptions{
			StartTime: time.Now().UTC().Add(-time.Hour), // 1 hour ago
			EndTime:   time.Now().UTC(),                 // Now
			Limit:     10000,
		})

		if err != nil {
			writeDataStoreError(w, err)
			return
		}
		writeJSON(w, logs)
	})
}
