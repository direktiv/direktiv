package core

import (
	"github.com/direktiv/direktiv/pkg/refactor/registry"
	"github.com/direktiv/direktiv/pkg/refactor/service"
)

type Config struct {
	ApiV1Port int `env:"API_V1_PORT" envDefault:"6665"`
	ApiV2Port int `env:"API_V2_PORT" envDefault:"6667"`
	GrpcPort  int `env:"API_PORT" envDefault:"6666"`

	Prometheus     string `env:"PROMETHEUS_BACKEND"`
	OpenTelemetry  string `env:"OPEN_TELEMETRY_BACKEND"`
	EnableEventing bool   `env:"ENABLE_EVENTING"`
}

type Version struct {
	UnixTime int64 `json:"unix_time"`
}

type App struct {
	Version         *Version
	ServiceManager  *service.Manager
	RegistryManager registry.Manager
}
