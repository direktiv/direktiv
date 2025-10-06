package api

import (
	"net/http"
	"path/filepath"

	"github.com/direktiv/direktiv/internal/core"
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

	workflowPath := r.URL.Query().Get("workflowPath")
	if workflowPath != "" {
		if workflowPath != filepath.Clean(workflowPath) {
			writeError(w, &Error{
				Code:    "request_invalid_parm",
				Message: "invalid request `workflowPath` param",
			})

			return
		}
		workflowPath = filepath.Clean(workflowPath)
		workflowPath = filepath.Join("/", workflowPath)
	}

	list, _, err := e.engine.GetInstances(r.Context(), ns, 0, 0)
	if err != nil {
		writeEngineError(w, err)

		return
	}

	allStatuses := []string{
		"pending", "failed", "complete", "cancelled", "crashed",
	}
	stats := make(map[string]int)
	stats["total"] = 0

	for _, s := range allStatuses {
		stats[s] = 0
	}

	foundMatching := false
	for _, v := range list {
		if workflowPath != "" {
			if v.Metadata[core.EngineMappingPath] == workflowPath {
				foundMatching = true
			}
			if v.Metadata[core.EngineMappingPath] != workflowPath {
				continue
			}
		}
		stats[v.StatusString()]++
		stats["total"]++
	}
	if !foundMatching && workflowPath != "" {
		writeError(w, &Error{
			Code:    "resource_not_found",
			Message: "requested workflow is not found",
		})

		return
	}

	writeJSON(w, stats)
}
