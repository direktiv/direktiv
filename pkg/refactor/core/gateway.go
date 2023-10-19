package core

import (
	"net/http"
)

type GatewayManager interface {
	ListEndpoints() []*Endpoint
	SetEndpoints(endpoints []*Endpoint)
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type Endpoint struct {
	DirektivAPI    string   `json:"direktiv_api"`
	Path           string   `json:"path"`
	Method         string   `json:"method"`
	TargetPlugin   Plugin   `json:"target_plugin"`
	TimeoutSeconds int      `json:"timeout_seconds"`
	AuthPlugins    []Plugin `json:"auth_plugins"`
	RequestPlugins []Plugin `json:"request_plugins"`
}

type Plugin struct {
	Name          string      `json:"name"`
	Version       string      `json:"version"`
	RuntimeConfig interface{} `json:"runtime_config"`
}

type GetPluginSchema func(key string) (string, bool)
