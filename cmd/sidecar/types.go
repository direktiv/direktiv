package sidecar

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type createVarRequest struct {
	Name             string `json:"name"`
	MimeType         string `json:"mimeType"`
	Data             []byte `json:"data"`
	InstanceIDString string `json:"instanceId"`
	WorkflowPath     string `json:"workflowPath"`
	Error            Error  `json:"error"`
}

type Error struct {
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Validation map[string]string `json:"validation,omitempty"`
}

type apiError struct {
	Error Error `json:"error"`
}

type RessourceNotFoundError struct {
	Key   string
	Scope string
}

func (e *RessourceNotFoundError) Error() string {
	return fmt.Sprintf("ressource with key %s not found in scope %s", e.Key, e.Scope)
}

type variable struct {
	ID        uuid.UUID `json:"id"`
	Typ       string    `json:"type"`
	Reference string    `json:"reference"`
	Name      string    `json:"name"`
	Data      []byte    `json:"data"`
	Size      int       `json:"size"`
	MimeType  string    `json:"mimeType"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
type variablesResponse struct {
	Data  []variable `json:"data"`
	Error *any       `json:"error"`
}

type variableResponse struct {
	Data  variable `json:"data"`
	Error *any     `json:"error"`
}

type file struct {
	Path      string      `json:"path"`
	Type      string      `json:"type"`
	Data      string      `json:"data"`
	Size      int         `json:"size"`
	MIMEType  string      `json:"mimeType"`
	CreatedAt string      `json:"createdAt"`
	UpdatedAt string      `json:"updatedAt"`
	Children  interface{} `json:"children,omitempty"`
}

type decodedFilesResponse struct {
	Error any  `json:"error,omitempty"`
	Data  file `json:"data,omitempty"`
}
