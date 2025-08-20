package api

import (
	"io"
	"net/http"
	"time"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/telemetry"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
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
	EndedAt      *time.Time     `json:"endedAt"`
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

func marshalForAPIJS(ins *datastore.JSInstance) *InstanceData {
	resp := &InstanceData{
		ID:           ins.ID,
		CreatedAt:    ins.CreatedAt,
		EndedAt:      nil,
		Status:       ins.StatusString(),
		WorkflowPath: ins.WorkflowPath,
		Invoker:      "test",
		Definition:   []byte(ins.WorkflowData),
		Namespace:    ins.Namespace,
	}

	return resp
}

type instController struct {
	db       *database.DB
	manager  any
	jsEngine core.JSEngine
}

func (e *instController) mountRouter(r chi.Router) {
	r.Get("/{instanceID}/subscribe", e.dummy)

	r.Get("/{instanceID}/input", e.dummy)
	r.Get("/{instanceID}/output", e.dummy)
	r.Get("/{instanceID}/metadata", e.dummy)

	r.Get("/{instanceID}", e.dummy)
	r.Patch("/{instanceID}", e.dummy)

	r.Get("/", e.dummy)
	r.Post("/", e.dummy)
}

func (e *instController) dummy(w http.ResponseWriter, r *http.Request) {
}

func (e *instController) create(w http.ResponseWriter, r *http.Request) {
	ns := extractContextNamespace(r)
	path := r.URL.Query().Get("path")

	wait := r.URL.Query().Get("wait") == "true"

	ctx := telemetry.GetContextFromRequest(r.Context(), r)
	ctx, span := telemetry.Tracer.Start(ctx, "api-request")
	span.SetAttributes(
		attribute.KeyValue{
			Key:   "namespace",
			Value: attribute.StringValue(ns.Name),
		},
		attribute.KeyValue{
			Key:   "path",
			Value: attribute.StringValue(path),
		},
		attribute.KeyValue{
			Key:   "wait",
			Value: attribute.BoolValue(wait),
		},
	)
	defer span.End()

	input, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}

	if wait && len(input) == 0 {
		input = []byte(`{}`)
	}

	id, err := e.jsEngine.ExecWorkflow(ctx, ns.Name, path, string(input))
	if err != nil {
		// telemetry.ReportError(span, err)
		writeError(w, &Error{
			Code:    err.Error(),
			Message: err.Error(),
		})

		return
	}

	data, err := e.db.DataStore().JSInstances().GetByID(ctx, id)
	if err != nil {
		writeError(w, &Error{
			Code:    err.Error(),
			Message: err.Error(),
		})

		return
	}

	writeJSON(w, marshalForAPIJS(data))
}
