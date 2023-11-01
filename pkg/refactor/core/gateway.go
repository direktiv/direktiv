package core

import (
	"net/http"
)

const MagicalGatewayNamespace = "gateway_namespace"

type EndpointManager interface {
	http.Handler

	GetAll() []*EndpointStatus
	SetEndpoints(endpoints []*Endpoint)
}

type Endpoint struct {
	Method   string   `json:"method"`
	FilePath string   `json:"file_path"`
	Plugins  []Plugin `json:"plugins"`
}
type Plugin struct {
	Type          string                 `json:"type"`
	Configuration map[string]interface{} `json:"configuration"`
}

type EndpointStatus struct {
	Endpoint
	Error string `json:"error"`
}
