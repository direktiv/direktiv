package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/direktiv/direktiv/internal/engine"
	"github.com/direktiv/direktiv/internal/sched"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LineageData struct {
	Branch int    `json:"branch"`
	ID     string `json:"id"`
	State  string `json:"state"`
	Step   int    `json:"step"`
}

type InstanceData struct {
	ID           uuid.UUID      `json:"id"`
	CreatedAt    time.Time      `json:"createdAt"`
	Started      time.Time      `json:"startedAt"`
	EndedAt      time.Time      `json:"endedAt"`
	Status       string         `json:"status"`
	WorkflowPath string         `json:"path"`
	ErrorCode    *string        `json:"errorCode"`
	Invoker      string         `json:"invoker"`
	Definition   []byte         `json:"definition,omitempty"`
	ErrorMessage []byte         `json:"errorMessage"`
	Flow         []string       `json:"flow"`
	TraceID      string         `json:"traceId"`
	Lineage      []*LineageData `json:"lineage"`
	Namespace    string         `json:"namespace"`

	InputLength    *int   `json:"inputLength,omitempty"`
	Input          []byte `json:"input,omitempty"`
	OutputLength   *int   `json:"outputLength,omitempty"`
	Output         []byte `json:"output,omitempty"`
	MetadataLength *int   `json:"metadataLength,omitempty"`
	Metadata       []byte `json:"metadata,omitempty"`
}

type instController struct {
	db        *gorm.DB
	manager   any
	engine    *engine.Engine
	scheduler *sched.Scheduler
}

func marshalForAPI(data *engine.InstanceStatus) (*InstanceData, error) {
	resp := &InstanceData{
		ID:           data.InstanceID,
		CreatedAt:    data.CreatedAt,
		Started:      data.StartedAt,
		EndedAt:      data.EndedAt,
		Status:       data.StatusString(),
		WorkflowPath: data.Metadata["workflowPath"],
		ErrorCode:    nil,
		Invoker:      "api",
		Definition:   []byte(data.Script),
		ErrorMessage: nil,
		Flow:         []string{},
		TraceID:      "",
		Lineage:      []*LineageData{},
		Namespace:    data.Namespace,
	}

	return resp, nil
}

func (e *instController) mountRouter(r chi.Router) {
	r.Get("/{instanceID}/subscribe", e.dummy)

	r.Get("/{instanceID}/input", e.dummy)
	r.Get("/{instanceID}/output", e.dummy)
	r.Get("/{instanceID}/metadata", e.dummy)

	r.Patch("/{instanceID}", e.dummy)

	r.Get("/", e.list)
	r.Get("/{instanceID}", e.get)

	r.Post("/", e.create)
	r.Get("/stats", e.stats)

	r.Post("/sched", e.createSched)
	r.Get("/sched", e.listSched)
}

func (e *instController) dummy(w http.ResponseWriter, r *http.Request) {
}

func (e *instController) create(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")
	path := r.URL.Query().Get("path")

	input, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}

	id, notify, err := e.engine.RunWorkflow(r.Context(), namespace, path, string(input), map[string]string{
		"workflowPath": path,
	})
	if err != nil {
		writeError(w, &Error{
			Code:    err.Error(),
			Message: err.Error(),
		})

		return
	}

	// var data *engine.InstanceStatus
	// for range 10 {
	// 	data, err = e.engine.GetInstanceByID(r.Context(), namespace, id)
	// 	if err != nil && errors.Is(err, engine.ErrDataNotFound) {
	// 		time.Sleep(5 * time.Millisecond)
	// 		continue
	// 	}
	// 	if err != nil {
	// 		writeError(w, &Error{
	// 			Code:    err.Error(),
	// 			Message: err.Error(),
	// 		})

	// 		return
	// 	}
	// }

	<-notify
	writeJSON(w, id)
}

func (e *instController) stats(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")

	list, err := e.engine.GetInstances(r.Context(), namespace, 0, 0)
	if err != nil {
		writeError(w, &Error{
			Code:    err.Error(),
			Message: err.Error(),
		})

		return
	}
	stats := make(map[string]int)
	for i := range list {
		n, ok := stats[list[i].Status]
		if !ok {
			stats[list[i].Status] = 0
		}
		stats[list[i].Status] = n + 1
	}

	writeJSON(w, stats)
}

func (e instController) testTranspile(w http.ResponseWriter, r *http.Request) {
	// namespace := chi.URLParam(r, "namespace")
	// path := r.URL.Query().Get("path")

	// if path == "" {
	// 	writeError(w, &Error{
	// 		Code:    "resource_not_found",
	// 		Message: "path not provided",
	// 	})

	// 	return
	// }

	// compiler, err := compiler.NewCompiler(e.db)
	// if err != nil {
	// 	writeInternalError(w, err)
	// 	return
	// }

	// flow, err := compiler.Compile(r.Context(), namespace, path)
	// if err != nil {
	// 	writeInternalError(w, err)
	// 	return
	// }

	// fmt.Printf("%+v\n", flow)

	// wait := r.URL.Query().Get("wait") == "true"
	// output := r.URL.Query().Get("output") == "true"

	// ctx := telemetry.GetContextFromRequest(r.Context(), r)
	// ctx, span := telemetry.Tracer.Start(ctx, "api-request")
	// span.SetAttributes(
	// 	attribute.KeyValue{
	// 		Key:   "namespace",
	// 		Value: attribute.StringValue(ns.Name),
	// 	},
	// 	attribute.KeyValue{
	// 		Key:   "path",
	// 		Value: attribute.StringValue(path),
	// 	},
	// 	attribute.KeyValue{
	// 		Key:   "wait",
	// 		Value: attribute.BoolValue(wait),
	// 	},
	// )
	// defer span.End()

	// input, err := io.ReadAll(r.Body)
	// if err != nil {
	// 	return
	// }

	// if wait && len(input) == 0 {
	// 	input = []byte(`{}`)
	// }

	// data, err := e.manager.Start(ctx, ns.Name, path, input)
	// if err != nil {
	// 	// telemetry.ReportError(span, err)
	// 	writeError(w, &Error{
	// 		Code:    err.Error(),
	// 		Message: err.Error(),
	// 	})

	// 	return
	// }

	// if wait {
	// 	e.handleWait(ctx, w, r, data, output)

	// 	return
	// }

	// writeJSON(w, marshalForAPI(data))
}

func (e *instController) get(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")
	instanceIDStr := chi.URLParam(r, "instanceID")
	instanceID, err := uuid.Parse(instanceIDStr)
	if err != nil {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: fmt.Errorf("unparsable instance UUID: %w", err).Error(),
		})

		return
	}

	data, err := e.engine.GetInstanceByID(r.Context(), namespace, instanceID)
	if err != nil {
		writeInternalError(w, err)

		return
	}

	resp, err := marshalForAPI(data)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, resp)
}

func (e *instController) list(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")

	limit := parseNumberQueryParam(r, "limit")
	offset := parseNumberQueryParam(r, "offset")

	list, err := e.engine.GetInstances(r.Context(), namespace, limit, offset)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	out := make([]any, len(list))
	for i := range list {
		obj, err := marshalForAPI(list[i])
		if err != nil {
			writeInternalError(w, err)
			return
		}
		out[i] = obj
	}

	metaInfo := map[string]any{
		"total": len(out),
	}

	writeJSONWithMeta(w, out, metaInfo)
}

// TODO: remove this test-code.
func (e *instController) createSched(w http.ResponseWriter, r *http.Request) {
	cfg := &sched.Rule{}
	if err := json.NewDecoder(r.Body).Decode(cfg); err != nil {
		writeError(w, &Error{
			Code:    "error",
			Message: err.Error(),
		})

		return
	}

	cfg, err := e.scheduler.SetRule(r.Context(), cfg)
	if err != nil {
		writeError(w, &Error{
			Code:    "error",
			Message: err.Error(),
		})

		return
	}

	writeJSON(w, cfg)
}

// TODO: remove this test-code.
func (e *instController) listSched(w http.ResponseWriter, r *http.Request) {
	list, err := e.scheduler.ListRules(r.Context())
	if err != nil {
		writeError(w, &Error{
			Code:    "error",
			Message: err.Error(),
		})

		return
	}

	writeJSON(w, list)
}

func parseNumberQueryParam(r *http.Request, name string) int {
	x := r.URL.Query().Get(name)
	if x == "" {
		return 0
	}
	k, err := strconv.ParseInt(x, 10, 64)
	if err != nil {
		return 0
	}

	return int(k)
}
