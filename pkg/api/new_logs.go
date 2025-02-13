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
		start := time.Now().UTC().Add(-time.Hour)
		end := time.Now().UTC()
		logs, err := m.meta.Get(r.Context(), metastore.LogQueryOptions{
			StartTime: &start, // 1 hour ago
			EndTime:   &end,   // Now
			Metadata: map[string]string{
				"namespace": "test",
				"instance":  "8f129335-9fb6-4353-989f-4d34c78f5b1b",
			},
			Limit: 10000,
		})

		if err != nil {
			writeDataStoreError(w, err)
			return
		}
		writeJSON(w, logs)
	})
}
