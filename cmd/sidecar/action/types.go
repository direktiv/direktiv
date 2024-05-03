package action

// TODO maybe move to flow.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ActionDeserialize interface {
	Extract(r *http.Request) (string, ActionRequest, error)
}

type ActionController struct {
	ActionRequest
	Cancel func()
}

type ActionRequest struct {
	Deadline  time.Duration            `json:"deadline"`
	UserInput []byte                   `json:"userInput"`
	Files     []FunctionFileDefinition `json:"files"`
	ActionContext
}

type ActionContext struct {
	Trace     string `json:"trace"`
	Span      string `json:"span"`
	State     string `json:"state"`
	Branch    string `json:"branch"`
	Instance  string `json:"instance"`
	Workflow  string `json:"workflow"`
	Namespace string `json:"namespace"`
	Callpath  string `json:"callpath"`
}

type FunctionFileDefinition struct {
	Key         string `json:"key"`
	As          string `json:"as,omitempty"`
	Scope       string `json:"scope,omitempty"`
	Type        string `json:"type,omitempty"`
	Permissions string `json:"permissions,omitempty"`
	Content     string `json:"content,omitempty"`
}

type ActionResponse struct {
	UserOutput []byte `json:"userOutput"`
	Err        any    `json:"err"`
	ErrCode    string `json:"errCode"`
}

type ActionRequestBuilder struct{}

func (ActionRequestBuilder) Extract(r *http.Request) (string, ActionRequest, error) {
	// 1. Retrieve necessary data from the context
	var c ActionRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&c); err != nil {
		return "", ActionRequest{}, fmt.Errorf("error reading request body: %w", err)
	}

	actionID := r.URL.Query().Get("action_id")
	if actionID == "" {
		return "", ActionRequest{}, fmt.Errorf("missing action_id in query parameters")
	}

	return actionID, c, nil
}
