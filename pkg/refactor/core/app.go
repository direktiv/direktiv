package core

import (
	"github.com/direktiv/direktiv/pkg/refactor/registry"
	"github.com/direktiv/direktiv/pkg/refactor/service"
)

type Config struct {
	ApiV1Port int `env:"DIREKTIV_API_V1_PORT" envDefault:"6665"`
	ApiV2Port int `env:"DIREKTIV_API_V2_PORT" envDefault:"6667"`
	GrpcPort  int `env:"DIREKTIV_GRPC_PORT" envDefault:"6666"`

	Prometheus     string `env:"DIREKTIV_PROMETHEUS_BACKEND"`
	OpenTelemetry  string `env:"DIREKTIV_OPEN_TELEMETRY_BACKEND"`
	EnableEventing bool   `env:"DIREKTIV_ENABLE_EVENTING"`
}

type Version struct {
	UnixTime int64 `json:"unix_time"`
}

type App struct {
	Version         *Version
	ServiceManager  *service.Manager
	RegistryManager registry.Manager
}
