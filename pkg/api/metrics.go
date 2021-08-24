package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

func (h *Handler) queryPrometheus(str string, t time.Time) (map[string]interface{}, error) {

	v1API := v1.NewAPI(h.s.prometheus)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, warnings, err := v1API.Query(ctx, str, t)
	if err != nil {
		return nil, err
	}

	out := map[string]interface{}{
		"warnings": warnings,
		"results":  res,
	}

	return out, nil
}

func (h *Handler) getNamespaceMetrics_WorkflowsInvoked(w http.ResponseWriter, r *http.Request) {

	out, err := h.queryPrometheus(fmt.Sprintf(`direktiv_workflows_invoked_total{namespace="%s"}`, mux.Vars(r)["namespace"]), time.Now())
	if err != nil {
		ErrResponse(w, err)
		return
	}

	writeJSONResponse(w, out)
}

func (h *Handler) getNamespaceMetrics_WorkflowsSuccessful(w http.ResponseWriter, r *http.Request) {

	out, err := h.queryPrometheus(fmt.Sprintf(`direktiv_workflows_success_total{namespace="%s"}`, mux.Vars(r)["namespace"]), time.Now())
	if err != nil {
		ErrResponse(w, err)
		return
	}

	writeJSONResponse(w, out)
}

func (h *Handler) getNamespaceMetrics_WorkflowsFailed(w http.ResponseWriter, r *http.Request) {

	out, err := h.queryPrometheus(fmt.Sprintf(`direktiv_workflows_failed_total{namespace="%s"}`, mux.Vars(r)["namespace"]), time.Now())
	if err != nil {
		ErrResponse(w, err)
		return
	}

	writeJSONResponse(w, out)
}

func (h *Handler) getNamespaceMetrics_WorkflowsMilliseconds(w http.ResponseWriter, r *http.Request) {

	out, err := h.queryPrometheus(fmt.Sprintf(`direktiv_workflows_total_milliseconds_sum{namespace="%s"}`, mux.Vars(r)["namespace"]), time.Now())
	if err != nil {
		ErrResponse(w, err)
		return
	}

	writeJSONResponse(w, out)
}

func (h *Handler) getWorkflowMetrics_Invoked(w http.ResponseWriter, r *http.Request) {
	out, err := h.queryPrometheus(fmt.Sprintf(`direktiv_workflows_invoked_total{namespace="%s", workflow="%s"}`, mux.Vars(r)["namespace"], mux.Vars(r)["workflow"]), time.Now())
	if err != nil {
		ErrResponse(w, err)
		return
	}

	writeJSONResponse(w, out)
}

func (h *Handler) getWorkflowMetrics_Successful(w http.ResponseWriter, r *http.Request) {
	out, err := h.queryPrometheus(fmt.Sprintf(`direktiv_workflows_success_total{namespace="%s", workflow="%s"}`, mux.Vars(r)["namespace"], mux.Vars(r)["workflow"]), time.Now())
	if err != nil {
		ErrResponse(w, err)
		return
	}

	writeJSONResponse(w, out)
}

func (h *Handler) getWorkflowMetrics_Failed(w http.ResponseWriter, r *http.Request) {
	out, err := h.queryPrometheus(fmt.Sprintf(`direktiv_workflows_failed_total{namespace="%s", workflow="%s"}`, mux.Vars(r)["namespace"], mux.Vars(r)["workflow"]), time.Now())
	if err != nil {
		ErrResponse(w, err)
		return
	}

	writeJSONResponse(w, out)
}

func (h *Handler) getWorkflowMetrics_Milliseconds(w http.ResponseWriter, r *http.Request) {

	out, err := h.queryPrometheus(fmt.Sprintf(`direktiv_workflows_total_milliseconds_sum{namespace="%s", workflow="%s"}`, mux.Vars(r)["namespace"], mux.Vars(r)["workflow"]), time.Now())
	if err != nil {
		ErrResponse(w, err)
		return
	}

	writeJSONResponse(w, out)
}
