package core

import (
	"net/http"
)

const MagicalGatewayNamespace = "gateway_namespace"

type EndpointManager interface {
	http.Handler

	GetAll() []*Endpoint
	SetEndpoints([]*Endpoint)
}

type Endpoint struct {
	Methods       []string `json:"methods"`
	FilePath      string   `json:"file_path"`
	PathExtension string   `yaml:"path-extension"`
	Plugins       []Plugin `json:"plugins"`
	Error         string   `json:"error"`
}

type Consumer struct {
	Username string   `yaml:"username"`
	Password string   `yaml:"password"`
	APIkey   string   `yaml:"apikey"`
	Tags     []string `yaml:"tags"`
	Groups   []string `yaml:"groups"`
}

type Plugin struct {
	Type          string                 `json:"type"`
	Configuration map[string]interface{} `json:"configuration"`
}
