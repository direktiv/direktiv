package api

import (
	"net/http"
	"path/filepath"

	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/go-chi/chi/v5"
)

type metricsController struct {
	db *database.SQLStore
}

func (e *metricsController) mountRouter(r chi.Router) {
	r.Get("/instances", e.instances)
}

func (e *metricsController) instances(w http.ResponseWriter, r *http.Request) {
	ns := extractContextNamespace(r)

	forWorkflowPath := r.URL.Query().Get("workflowPath")
	if forWorkflowPath != "" && forWorkflowPath != filepath.Clean(forWorkflowPath) {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: "query param workflowPath invalid file path",
		})

		return
	}

	db, err := e.db.BeginTx(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	defer db.Rollback()

	counts, err := db.InstanceStore().GetNamespaceInstanceCounts(r.Context(), ns.ID, forWorkflowPath)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, map[string]interface{}{
		"complete":  counts.Complete,
		"failed":    counts.Failed,
		"crashed":   counts.Crashed,
		"cancelled": counts.Cancelled,
		"pending":   counts.Pending,
		"total":     counts.Total,
	})
}
