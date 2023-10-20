package core

import (
	"net/http"
	"sync"
)

const MagicalGatewayNamespace = "gateway_namespace"

type GatewayManager interface {
	http.Handler

	ListEndpoints() []*Endpoint
	SetEndpoints(endpoints []*Endpoint)

	Start(done <-chan struct{}, wg *sync.WaitGroup)
}

type Endpoint struct {
	Method string `json:"method"`
}
