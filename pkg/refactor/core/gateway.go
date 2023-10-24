package core

import (
	"net/http"
)

const MagicalGatewayNamespace = "gateway_namespace"

type EndpointManager interface {
	http.Handler

	GetAll() []*EndpointStatus
	SetEndpoints(endpoints []*Endpoint) []*EndpointStatus
}

type Endpoint struct {
	Method    string    `json:"method"`
	FilePath  string    `json:"file_path"`
	Workflow  string    `json:"workflow"`
	Namespace string    `json:"namespace"`
	Plugins   []Plugins `json:"plugins"`
}
type Plugins struct {
	Type          string      `json:"type"`
	Configuration interface{} `json:"configuration"`
}

type EndpointStatus struct {
	Endpoint
	Status string `json:"status"`
	Error  string `json:"error"`
}
