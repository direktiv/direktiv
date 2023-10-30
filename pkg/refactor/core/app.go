package core

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("ErrNotFound")

// nolint:revive,stylecheck
type Config struct {
	ApiV1Port int `env:"DIREKTIV_API_V1_PORT" envDefault:"6665"`
	ApiV2Port int `env:"DIREKTIV_API_V2_PORT" envDefault:"6667"`
	GrpcPort  int `env:"DIREKTIV_GRPC_PORT"   envDefault:"6666"`

	Prometheus     string `env:"DIREKTIV_PROMETHEUS_BACKEND"`
	OpenTelemetry  string `env:"DIREKTIV_OPEN_TELEMETRY_BACKEND"`
	EnableEventing bool   `env:"DIREKTIV_ENABLE_EVENTING"`

	SecretKey string `env:"DIREKTIV_SECRET_KEY" envDefault:"01234567890123456789012345678912"`

	DB string `env:"DIREKTIV_DB,notEmpty"`

	KnativeServiceAccount string `env:"DIREKTIV_KNATIVE_SERVICE_ACCOUNT"`
	KnativeNamespace      string `env:"DIREKTIV_KNATIVE_NAMESPACE"`
	KnativeIngressClass   string `env:"DIREKTIV_KNATIVE_INGRESS_CLASS"`
	KnativeSidecar        string `env:"DIREKTIV_KNATIVE_SIDECAR"`
	KnativeMaxScale       int    `env:"DIREKTIV_KNATIVE_MAX_SCALE"   envDefault:"5"`
	KnativeNetShape       string `env:"DIREKTIV_KNATIVE_NET_SHAPE"`

	FunctionsTimeout int    `env:"DIREKTIV_FUNCTIONS_TIMEOUT" envDefault:"7200"`
	LogFormat        string `env:"DIREKTIV_LOG_JSON"`
	LogDebug         bool   `env:"DIREKTIV_DEBUG"`
}

func (conf *Config) GetFunctionsTimeout() time.Duration {
	return time.Second * time.Duration(conf.FunctionsTimeout)
}

type Version struct {
	UnixTime int64 `json:"unix_time"`
}

type App struct {
	Version *Version
	Config  *Config

	ServiceManager  ServiceManager
	RegistryManager RegistryManager
	GatewayManager  GatewayManager
}
