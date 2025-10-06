package api

import (
	"io"
	"net/http"
	"path/filepath"
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

	InputLength    int    `json:"inputLength"`
	Input          []byte `json:"input"`
	OutputLength   int    `json:"outputLength"`
	Output         []byte `json:"output"`
	MetadataLength int    `json:"metadataLength"`
	Metadata       []byte `json:"metadata"`
}

type instController struct {
	db        *gorm.DB
	manager   any
	engine    *engine.Engine
	scheduler *sched.Scheduler
}

func convertInstanceData(data *engine.InstanceStatus) *InstanceData {
	resp := &InstanceData{
		ID:             data.InstanceID,
		CreatedAt:      data.CreatedAt,
		Started:        data.StartedAt,
		EndedAt:        data.EndedAt,
		Status:         data.StatusString(),
		WorkflowPath:   data.Metadata["workflowPath"],
		ErrorCode:      nil,
		Invoker:        "api",
		Definition:     []byte(data.Script),
		ErrorMessage:   nil,
		Flow:           []string{},
		TraceID:        "",
		Lineage:        []*LineageData{},
		Namespace:      data.Namespace,
		InputLength:    len(data.Input),
		OutputLength:   len(data.Output),
		MetadataLength: len(data.Metadata),
		Input:          data.Input,
		Output:         data.Output,
	}
	if data.Error != "" {
		resp.ErrorMessage = []byte(data.Error)
	}

	return resp
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
}

func (e *instController) dummy(w http.ResponseWriter, r *http.Request) {
}

func (e *instController) create(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")
	path := r.URL.Query().Get("path")
	if path != filepath.Clean(path) {
		writeError(w, &Error{
			Code:    "request_invalid_param",
			Message: "invalid request `path` param",
		})
	}
	path = filepath.Clean(path)
	path = filepath.Join("/", path)

	input, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, &Error{
			Code:    "request_invalid_data",
			Message: "could not read request body",
		})

		return
	}

	_, notify, err := e.engine.RunWorkflow(r.Context(), namespace, path, string(input), map[string]string{
		"workflowPath": path,
	})
	if err != nil {
		writeEngineError(w, err)

		return
	}

	status := <-notify
	writeJSON(w, convertInstanceData(status))
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

	writeJSON(w, convertInstanceData(data))
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
		out[i] = convertInstanceData(list[i])
	}

	metaInfo := map[string]any{
		"total": total,
	}

	writeJSONWithMeta(w, out, metaInfo)
}
