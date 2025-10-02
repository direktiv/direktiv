package api

import (
	"net/http"
	"path/filepath"

	"github.com/direktiv/direktiv/internal/engine"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type metricsController struct {
	db     *gorm.DB
	engine *engine.Engine
}

func (e *metricsController) mountRouter(r chi.Router) {
	r.Get("/instances", e.instances)
}

// calculates the stats Status->Count of instances in the namespace.
func (e *metricsController) instances(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "namespace")

	forWorkflowPath := r.URL.Query().Get("workflowPath")
	if forWorkflowPath != "" && forWorkflowPath != filepath.Clean(forWorkflowPath) {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: "query param workflowPath invalid file path",
		})

		return
	}
	list, _, err := e.engine.GetInstances(r.Context(), ns, 0, 0)
	if err != nil {
		writeEngineError(w, err)

		return
	}

	stats := make(map[string]int)
	stats["total"] = 0
	for _, v := range list {
		if forWorkflowPath != "" && v.Metadata["workflowPath"] != forWorkflowPath {
			continue
		}
		n, ok := stats[v.StatusString()]
		if !ok {
			stats[v.StatusString()] = 0
		}
		stats[v.StatusString()] = n + 1
		stats["total"]++
	}

	writeJSON(w, stats)
}
