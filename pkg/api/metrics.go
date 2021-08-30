package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

func (h *Handler) queryPrometheus(str string, t time.Time) (map[string]interface{}, error) {

	if !h.s.prometheusEnabled {
		return nil, fmt.Errorf("missing prometheus configuration")
	}

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

func (h *Handler) getWorkflowMetrics_StateMilliseconds(w http.ResponseWriter, r *http.Request) {

	ns := mux.Vars(r)["namespace"]
	wf := mux.Vars(r)["workflow"]

	out, err := h.queryPrometheus(fmt.Sprintf(`direktiv_states_milliseconds_sum{namespace="%s", workflow="%s"} / direktiv_states_milliseconds_count{namespace="%s", workflow="%s"}`, ns, wf, ns, wf), time.Now())
	if err != nil {
		ErrResponse(w, err)
		return
	}

	writeJSONResponse(w, out)
}

func intFromPrometheusResponse(res interface{}) (int, error) {

	m, ok := res.(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("unable to parse object")
	}

	if len(m) == 0 {
		return 0, nil
	}

	if xRes, ok := m["results"]; ok {

		b, err := json.Marshal(xRes)
		if err != nil {
			return 0, err
		}

		res := make([]interface{}, 0)
		err = json.Unmarshal(b, &res)
		if err != nil {
			return 0, err
		}

		if len(res) == 0 {
			return 0, nil
		}

		if xMap, ok := res[0].(map[string]interface{}); ok {
			if xVals, ok := xMap["value"]; ok {
				if vals, ok := xVals.([]interface{}); ok {
					if len(vals) > 1 {
						if str, ok := vals[1].(string); ok {

							n, err := strconv.Atoi(str)
							if err != nil {
								return 0, err
							}

							return n, nil
						}
					}
				}
			}
		}
	}

	return 0, nil
}

// satisfies the API call made by the UI to populate the 'Executed Workflows' panel on the UI
func (h *Handler) getWorkflowMetrics_Deprecated(w http.ResponseWriter, r *http.Request) {

	ns := mux.Vars(r)["namespace"]
	wf := mux.Vars(r)["workflow"]

	m1, err := h.queryPrometheus(fmt.Sprintf(`direktiv_workflows_invoked_total{namespace="%s", workflow="%s"}`, ns, wf), time.Now())
	if err != nil {
		ErrResponse(w, err)
		return
	}

	m2, err := h.queryPrometheus(fmt.Sprintf(`direktiv_workflows_invoked_total{namespace="%s", workflow="%s"}`, ns, wf), time.Now())
	if err != nil {
		ErrResponse(w, err)
		return
	}

	out := make(map[string]interface{})
	out["successfulExecutions"], err = intFromPrometheusResponse(m1)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	out["totalInstancesRun"], err = intFromPrometheusResponse(m2)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	writeJSONResponse(w, out)
}
