package sidecar

import (
	"log/slog"
	"sync"
)

// Config defines the configuration structure for environment variables.
type Config struct {
	InternalPort    string `env:"INTERNAL_PORT"`   // Port for the internal router.
	ExternalPort    string `env:"SIDECAR_PORT"`    // Port for the external router.
	FlowServerURL   string `env:"FLOW_SERVER_URL"` // Endpoint for forwarding task results.
	UserServiceURL  string `env:"USER_SERVICE_URL"`
	MaxResponseSize string `env:"MAX_RESPONSE_SIZE"`
}

func main() {
	var config Config
	var dataMap sync.Map // actionID -> Action
	slog.Debug("starting sidecar APIs")
	StartApis(config, &dataMap)
	slog.Debug("started sidecar APIs")
}
