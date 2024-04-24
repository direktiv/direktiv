package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

// GetInodePath returns the path without the first slash.
func GetInodePath(path string) string {
	path = strings.TrimSuffix(path, "/")
	if !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	path = filepath.Clean(path)
	return path
}

func (h *flowHandler) queryPrometheus(ctx context.Context, str string, t time.Time) (map[string]interface{}, error) {
	v1API := v1.NewAPI(h.prometheus)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
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

func (h *flowHandler) NamespaceMetricsInvoked(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "this", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]

	resp, err := h.queryPrometheus(ctx, fmt.Sprintf(`direktiv_workflows_invoked_total{direktiv_namespace="%s"}`, namespace), time.Now().UTC())
	respondJSON(w, resp, err)
}

func (h *flowHandler) WorkflowMetricsInvoked(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "this", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)
	path = GetInodePath(path)

	resp, err := h.queryPrometheus(ctx, fmt.Sprintf(`direktiv_workflows_invoked_total{direktiv_namespace="%s", direktiv_workflow="%s"}`, namespace, path), time.Now().UTC())
	respondJSON(w, resp, err)
}

func (h *flowHandler) NamespaceMetricsSuccessful(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "this", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]

	resp, err := h.queryPrometheus(ctx, fmt.Sprintf(`direktiv_workflows_success_total{direktiv_namespace="%s"}`, namespace), time.Now().UTC())
	respondJSON(w, resp, err)
}

func (h *flowHandler) WorkflowMetricsSuccessful(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "this", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)
	path = GetInodePath(path)

	resp, err := h.queryPrometheus(ctx, fmt.Sprintf(`direktiv_workflows_success_total{direktiv_namespace="%s", direktiv_workflow="%s"}`, namespace, path), time.Now().UTC())
	respondJSON(w, resp, err)
}

func (h *flowHandler) NamespaceMetricsFailed(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "this", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]

	resp, err := h.queryPrometheus(ctx, fmt.Sprintf(`direktiv_workflows_failed_total{direktiv_namespace="%s"}`, namespace), time.Now().UTC())
	respondJSON(w, resp, err)
}

func (h *flowHandler) WorkflowMetricsFailed(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "this", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)
	path = GetInodePath(path)

	resp, err := h.queryPrometheus(ctx, fmt.Sprintf(`direktiv_workflows_failed_total{direktiv_namespace="%s", direktiv_workflow="%s"}`, namespace, path), time.Now().UTC())
	respondJSON(w, resp, err)
}

func (h *flowHandler) NamespaceMetricsMilliseconds(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "this", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]

	resp, err := h.queryPrometheus(ctx, fmt.Sprintf(`direktiv_workflows_total_milliseconds_sum{direktiv_namespace="%s"}`, namespace), time.Now().UTC())
	respondJSON(w, resp, err)
}

func (h *flowHandler) WorkflowMetricsMilliseconds(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "this", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)
	path = GetInodePath(path)

	resp, err := h.queryPrometheus(ctx, fmt.Sprintf(`direktiv_workflows_total_milliseconds_sum{direktiv_namespace="%s", direktiv_workflow="%s"}`, namespace, path), time.Now().UTC())
	respondJSON(w, resp, err)
}

func (h *flowHandler) WorkflowMetricsStateMilliseconds(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "this", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)
	path = GetInodePath(path)

	resp, err := h.queryPrometheus(ctx, fmt.Sprintf(`direktiv_states_milliseconds_sum{direktiv_namespace="%s", direktiv_workflow="%s"} / direktiv_states_milliseconds_count{namespace="%s", workflow="%s"}`, namespace, path, namespace, path), time.Now().UTC())
	respondJSON(w, resp, err)
}
