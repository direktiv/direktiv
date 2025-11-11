package api

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/direktiv/direktiv/internal/api/filter"
	"github.com/direktiv/direktiv/internal/core"
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
	EndedAt      *time.Time     `json:"endedAt"`
	Status       string         `json:"status"`
	WorkflowPath string         `json:"path"`
	ErrorCode    *string        `json:"errorCode"`
	Invoker      string         `json:"invoker"`
	Definition   []byte         `json:"definition,omitempty"`
	ErrorMessage *string        `json:"errorMessage"`
	Flow         []string       `json:"flow"`
	TraceID      string         `json:"traceId"`
	Lineage      []*LineageData `json:"lineage"`
	Namespace    string         `json:"namespace"`

	InputLength    int     `json:"inputLength"`
	Input          string  `json:"input"`
	OutputLength   int     `json:"outputLength"`
	Output         *string `json:"output"`
	MetadataLength int     `json:"metadataLength"`
	Metadata       []byte  `json:"metadata"`
}

type InstanceEvent struct {
	State      string            `json:"state"`
	InstanceID uuid.UUID         `json:"instanceId"`
	Namespace  string            `json:"namespace"`
	Metadata   map[string]string `json:"metadata"`
	Script     string            `json:"script,omitempty"`
	Fn         string            `json:"fn,omitempty"`
	Mappings   string            `json:"mappings,omitempty"`

	Input  json.RawMessage `json:"input,omitempty"`
	Output json.RawMessage `json:"output,omitempty"`
	Error  string          `json:"error,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
	StartedAt time.Time `json:"startedAt"`
	EndedAt   time.Time `json:"endedAt"`

	EventID  uuid.UUID `json:"eventId"`
	Sequence uint64    `json:"sequence"`
}

func convertToInstanceEvent(data *engine.InstanceEvent) *InstanceEvent {
	return &InstanceEvent{
		State:      string(data.State),
		EventID:    data.EventID,
		InstanceID: data.InstanceID,
		Namespace:  data.Namespace,
		Metadata:   data.Metadata,
		Script:     data.Script,
		Fn:         data.Fn,
		Mappings:   data.Mappings,
		Input:      data.Input,
		Output:     data.Output,
		Error:      data.Error,
		CreatedAt:  data.CreatedAt,
		StartedAt:  data.StartedAt,
		EndedAt:    data.EndedAt,
		Sequence:   data.Sequence,
	}
}

func convertInstanceData(data *engine.InstanceEvent) *InstanceData {
	resp := &InstanceData{
		ID:             data.InstanceID,
		CreatedAt:      data.CreatedAt,
		Started:        data.StartedAt,
		Status:         string(data.State),
		WorkflowPath:   data.Metadata[core.EngineMappingPath],
		ErrorCode:      nil,
		Invoker:        "api",
		Definition:     []byte(data.Script),
		ErrorMessage:   nil,
		Flow:           []string{},
		TraceID:        "",
		Lineage:        []*LineageData{},
		Namespace:      data.Namespace,
		InputLength:    len(data.Input),
		Input:          string(data.Input),
		OutputLength:   len(data.Output),
		MetadataLength: len(data.Metadata),
	}
	if !data.EndedAt.IsZero() {
		resp.EndedAt = &data.EndedAt
	}
	if data.Output != nil {
		s := string(data.Output)
		resp.Output = &s
	}
	if data.Error != "" {
		resp.ErrorMessage = &data.Error
	}

	return resp
}

type instController struct {
	db        *gorm.DB
	manager   any
	engine    *engine.Engine
	scheduler *sched.Scheduler
}

func (e *instController) mountRouter(r chi.Router) {
	r.Get("/{instanceID}/subscribe", e.dummy)
	r.Get("/{instanceID}/input", e.dummy)
	r.Get("/{instanceID}/history", e.history)
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

	if len(input) == 0 {
		input = []byte("null")
	}

	withWait := r.URL.Query().Get("wait") == "true"
	withSyncExec := withWait
	if r.URL.Query().Get("asyncExec") == "true" {
		withSyncExec = false
	}

	st, notify, err := e.engine.StartWorkflow(r.Context(), namespace, path, string(input), map[string]string{
		core.EngineMappingPath:      path,
		core.EngineMappingNamespace: namespace,
		core.EngineMappingCaller:    "api",
		engine.LabelWithNotify:      strconv.FormatBool(withWait),
		engine.LabelWithSyncExec:    strconv.FormatBool(withSyncExec),
		engine.LabelInvokerType:  "api",
	})
	if err != nil {
		writeEngineError(w, err)

		return
	}

	if withWait {
		st = <-notify
	}

	writeJSON(w, convertInstanceData(st))
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

	data, err := e.engine.GetInstanceStatus(r.Context(), namespace, instanceID)
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

	queryValues := r.URL.Query()

	// We need to translate filter keys from the API to the engine.
	replacements := map[string]string{
		"[path]":    "[metadata_" + core.EngineMappingPath + "]",
		"[invoker]": "[metadata_" + engine.LabelInvokerType + "]",
	}
	fixedQueryValues := make(map[string][]string)
	for k, v := range queryValues {
		for rSrc, rDes := range replacements {
			k = strings.ReplaceAll(k, rSrc, rDes)
		}
		fixedQueryValues[k] = v
	}

	fil := filter.With(
		filter.FromURLValues(fixedQueryValues),
		filter.FieldEQ("namespace", namespace),
	)
	list, total, err := e.engine.ListInstanceStatuses(r.Context(), limit, offset, fil)
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

func (e *instController) history(w http.ResponseWriter, r *http.Request) {
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

	list, err := e.engine.GetInstanceHistory(r.Context(), namespace, instanceID)
	if err != nil {
		writeEngineError(w, err)
		return
	}
	out := make([]any, len(list))
	for i := range list {
		out[i] = convertToInstanceEvent(list[i])
	}

	writeJSON(w, out)
}
