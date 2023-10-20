package core

import (
	"net/http"
	"sync"
)

const MagicalGatewayNamespace = "gateway_namespace"

type EndpointManager interface {
	http.Handler

	GetAll() []*EndpointStatus
	SetEndpoints(endpoints []*Endpoint)

	Start(done <-chan struct{}, wg *sync.WaitGroup)
}

type Endpoint struct {
	Method    string `json:"method"`
	Workflow  string `json:"workflow"`
	Namespace string `json:"namespace"`
	Plugins   []struct {
		ID            string      `json:"id"`
		Configuration interface{} `json:"configuration"`
	} `json:"plugins"`
}

type EndpointStatus struct {
	Endpoint
	Status string `json:"status"`
	Error  string `json:"error"`
}
