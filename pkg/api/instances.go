package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/transpiler"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

type instController struct {
	db *database.DB
}

func (e *instController) mountRouter(r chi.Router) {
	r.Get("/{instanceID}/subscribe", e.dummy)

	r.Get("/{instanceID}/input", e.dummy)
	r.Get("/{instanceID}/output", e.dummy)
	r.Get("/{instanceID}/metadata", e.dummy)

	r.Get("/{instanceID}", e.dummy)
	r.Patch("/{instanceID}", e.dummy)

	r.Get("/", e.dummy)
	r.Post("/", e.execute)
}

func (e *instController) dummy(w http.ResponseWriter, r *http.Request) {
}

func (e instController) execute(w http.ResponseWriter, r *http.Request) {

	namespace := chi.URLParam(r, "namespace")
	path := r.URL.Query().Get("path")

	if path == "" {
		writeError(w, &Error{
			Code:    "resource_not_found",
			Message: "path not provided",
		})
		return
	}

	compiler, err := transpiler.NewCompiler(e.db)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	flow, err := compiler.Compile(r.Context(), namespace, path)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	fmt.Printf("%+v\n", flow)

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
