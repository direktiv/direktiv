package api

import (
	"io"
	"net/http"
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
}

func (e *instController) dummy(w http.ResponseWriter, r *http.Request) {
}

func (e *instController) create(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")
	path := r.URL.Query().Get("path")

	input, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: "invalid request body",
		})
		return
	}

	id, notify, err := e.engine.RunWorkflow(r.Context(), namespace, path, string(input), map[string]string{
		"workflowPath": path,
	})
	if err != nil {
		writeEngineError(w, err)

		return
	}

	<-notify
	writeJSON(w, id)
}

// calculates the stats Status->Count of all instances in the namespace
func (e *instController) stats(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")

	list, _, err := e.engine.GetInstances(r.Context(), namespace, 0, 0)
	if err != nil {
		writeEngineError(w, err)

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

func (e *instController) get(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")
	instanceIDStr := chi.URLParam(r, "instanceID")
	instanceID, err := uuid.Parse(instanceIDStr)
	if err != nil {
		writeError(w, &Error{
			Code:    "request_id_invalid",
			Message: "invalid instance uuid",
		})
		return
	}

	data, err := e.engine.GetInstanceByID(r.Context(), namespace, instanceID)
	if err != nil {
		writeEngineError(w, err)
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

	limit := ParseQueryParam[int](r, "limit", 0)
	offset := ParseQueryParam[int](r, "offset", 0)

	list, total, err := e.engine.GetInstances(r.Context(), namespace, limit, offset)
	if err != nil {
		writeEngineError(w, err)
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
		"total": total,
	}

	writeJSONWithMeta(w, out, metaInfo)
}
