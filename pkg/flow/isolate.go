package flow

import (
	"github.com/direktiv/direktiv/pkg/model"
)

const (
	DirektivActionIDHeader = "Direktiv-ActionID"
)

// ServiceResponse is the response structure for internal knative services.
type ServiceResponse struct {
	ErrorCode    string      `json:"errorCode"`
	ErrorMessage string      `json:"errorMessage"`
	Data         interface{} `json:"data"`
}

type functionRequest struct {
	ActionID  string
	Timeout   int
	Container functionContainer
}

type functionContainer struct {
	Type                model.FunctionType
	ID                  string
	Image, Cmd, Service string
	Data                []byte
	Size                model.Size
	Scale               int
	Files               []model.FunctionFileDefinition
}
