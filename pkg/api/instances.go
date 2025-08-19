package api

import (
	"net/http"
	"time"

	"github.com/direktiv/direktiv/pkg/database"
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
	db      *database.DB
	manager any
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
