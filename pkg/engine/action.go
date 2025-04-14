package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ActionRequest struct {
	// TODO secrets
	Deadline  time.Time                `json:"deadline"`
	UserInput []byte                   `json:"userInput"`
	Files     []FunctionFileDefinition `json:"files"`
	Async     bool                     `json:"async"`
	ActionContext
}

type ActionContext struct {
	TraceParent string `json:"traceParent"`
	State       string `json:"state"`
	Branch      int    `json:"branch"`
	Instance    string `json:"instance"`
	Workflow    string `json:"workflow"`
	Namespace   string `json:"namespace"`
	Action      string `json:"action"`
	Step        int    `json:"step"`
	Path        string `json:"path"`
	Invoker     string `json:"invoker"`
}

type FunctionFileDefinition struct {
	Key         string `json:"key"`
	As          string `json:"as,omitempty"`
	Scope       string `json:"scope,omitempty"`
	Type        string `json:"type,omitempty"`
	Permissions string `json:"permissions,omitempty"`
	Content     []byte `json:"content,omitempty"`
}

func DecodeActionRequest(r *http.Request) (ActionRequest, error) {
	var c ActionRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&c); err != nil {
		return ActionRequest{}, fmt.Errorf("error reading request body: %w", err)
	}

	return c, nil
}

func EncodeActionRequest(ar ActionRequest) (io.Reader, error) {
	encodedRequest, err := json.Marshal(ar)
	if err != nil {
		return nil, fmt.Errorf("error encoding response: %w", err)
	}

	return bytes.NewReader(encodedRequest), nil
}

type ActionResponse struct {
	Output  []byte `json:"output"`
	ErrMsg  string `json:"errMsg"`
	ErrCode string `json:"errCode"`
}
