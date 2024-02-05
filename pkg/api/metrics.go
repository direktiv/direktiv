package api

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/gorilla/mux"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
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

func (h *flowHandler) MetricsSankey(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)

	// QueryParams
	values := r.URL.Query()
	since := values.Get("since")

	var x time.Time
	if since != "" {
		dura, err := time.ParseDuration(since)
		if err != nil {
			code := http.StatusBadRequest
			http.Error(w, "invalid 'since' parameter", code)
			return
		}
		x = time.Now().UTC().Add(-1 * dura)
	}

	ts := timestamppb.New(x)

	in := &grpc.WorkflowMetricsRequest{
		Namespace:      namespace,
		Path:           path,
		SinceTimestamp: ts,
	}

	resp, err := h.client.WorkflowMetrics(ctx, in)
	respond(w, resp, err)
}

func (h *flowHandler) NamespaceMetricsInvoked(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]

	resp, err := h.queryPrometheus(ctx, fmt.Sprintf(`direktiv_workflows_invoked_total{direktiv_namespace="%s"}`, namespace), time.Now().UTC())
	respondJSON(w, resp, err)
}

func (h *flowHandler) WorkflowMetricsInvoked(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)
	path = GetInodePath(path)

	resp, err := h.queryPrometheus(ctx, fmt.Sprintf(`direktiv_workflows_invoked_total{direktiv_namespace="%s", direktiv_workflow="%s"}`, namespace, path), time.Now().UTC())
	respondJSON(w, resp, err)
}

func (h *flowHandler) NamespaceMetricsSuccessful(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]

	resp, err := h.queryPrometheus(ctx, fmt.Sprintf(`direktiv_workflows_success_total{direktiv_namespace="%s"}`, namespace), time.Now().UTC())
	respondJSON(w, resp, err)
}

func (h *flowHandler) WorkflowMetricsSuccessful(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)
	path = GetInodePath(path)

	resp, err := h.queryPrometheus(ctx, fmt.Sprintf(`direktiv_workflows_success_total{direktiv_namespace="%s", direktiv_workflow="%s"}`, namespace, path), time.Now().UTC())
	respondJSON(w, resp, err)
}

func (h *flowHandler) NamespaceMetricsFailed(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]

	resp, err := h.queryPrometheus(ctx, fmt.Sprintf(`direktiv_workflows_failed_total{direktiv_namespace="%s"}`, namespace), time.Now().UTC())
	respondJSON(w, resp, err)
}

func (h *flowHandler) WorkflowMetricsFailed(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)
	path = GetInodePath(path)

	resp, err := h.queryPrometheus(ctx, fmt.Sprintf(`direktiv_workflows_failed_total{direktiv_namespace="%s", direktiv_workflow="%s"}`, namespace, path), time.Now().UTC())
	respondJSON(w, resp, err)
}

func (h *flowHandler) NamespaceMetricsMilliseconds(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]

	resp, err := h.queryPrometheus(ctx, fmt.Sprintf(`direktiv_workflows_total_milliseconds_sum{direktiv_namespace="%s"}`, namespace), time.Now().UTC())
	respondJSON(w, resp, err)
}

func (h *flowHandler) WorkflowMetricsMilliseconds(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)
	path = GetInodePath(path)

	resp, err := h.queryPrometheus(ctx, fmt.Sprintf(`direktiv_workflows_total_milliseconds_sum{direktiv_namespace="%s", direktiv_workflow="%s"}`, namespace, path), time.Now().UTC())
	respondJSON(w, resp, err)
}

func (h *flowHandler) WorkflowMetricsStateMilliseconds(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Handling request: %s", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	path, _ := pathAndRef(r)
	path = GetInodePath(path)

	resp, err := h.queryPrometheus(ctx, fmt.Sprintf(`direktiv_states_milliseconds_sum{direktiv_namespace="%s", direktiv_workflow="%s"} / direktiv_states_milliseconds_count{namespace="%s", workflow="%s"}`, namespace, path, namespace, path), time.Now().UTC())
	respondJSON(w, resp, err)
}
