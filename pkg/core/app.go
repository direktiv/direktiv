package core

import (
	"errors"
	"fmt"
	"time"
)

var ErrNotFound = errors.New("ErrNotFound")

const SystemNamespace = "system"

// nolint:revive,stylecheck
type Config struct {
	DirektivNamespace string `env:"DIREKTIV_NAMESPACE"`

	LogDebug bool `env:"DIREKTIV_DEBUG"`

	ApiPort int `env:"DIREKTIV_API_PORT" envDefault:"6665"`

	SecretKey string `env:"DIREKTIV_SECRET_KEY,notEmpty"`
	DB        string `env:"DIREKTIV_DB,notEmpty"`

	FunctionsTimeout int `env:"DIREKTIV_FUNCTIONS_TIMEOUT" envDefault:"7200"`

	OpenTelemetry string `env:"DIREKTIV_OPEN_TELEMETRY_BACKEND"`

	KnativeServiceAccount string `env:"DIREKTIV_KNATIVE_SERVICE_ACCOUNT"`
	KnativeNamespace      string `env:"DIREKTIV_KNATIVE_NAMESPACE"`
	KnativeIngressClass   string `env:"DIREKTIV_KNATIVE_INGRESS_CLASS"`
	KnativeSidecar        string `env:"DIREKTIV_KNATIVE_SIDECAR"`
	KnativeMaxScale       int    `env:"DIREKTIV_KNATIVE_MAX_SCALE"       envDefault:"5"`
	KnativeNetShape       string `env:"DIREKTIV_KNATIVE_NET_SHAPE"`

	KnativeSizeMemorySmall  int `env:"DIREKTIV_KNATIVE_SIZE_MEMORY_SMALL"  envDefault:"512"`
	KnativeSizeMemoryMedium int `env:"DIREKTIV_KNATIVE_SIZE_MEMORY_MEDIUM" envDefault:"1024"`
	KnativeSizeMemoryLarge  int `env:"DIREKTIV_KNATIVE_SIZE_MEMORY_LARGE"  envDefault:"2048"`

	KnativeSizeCPUSmall  string `env:"DIREKTIV_KNATIVE_SIZE_CPU_SMALL"  envDefault:"250m"`
	KnativeSizeCPUMedium string `env:"DIREKTIV_KNATIVE_SIZE_CPU_MEDIUM" envDefault:"500m"`
	KnativeSizeCPULarge  string `env:"DIREKTIV_KNATIVE_SIZE_CPU_LARGE"  envDefault:"1"`

	KnativeSizeDiskSmall  int `env:"DIREKTIV_KNATIVE_SIZE_DISK_SMALL"  envDefault:"256"`
	KnativeSizeDiskMedium int `env:"DIREKTIV_KNATIVE_SIZE_DISK_MEDIUM" envDefault:"1024"`
	KnativeSizeDiskLarge  int `env:"DIREKTIV_KNATIVE_SIZE_DISK_LARGE"  envDefault:"4096"`

	KnativeProxyNo    string `env:"DIREKTIV_KNATIVE_PROXY_NO"`
	KnativeProxyHTTP  string `env:"DIREKTIV_KNATIVE_PROXY_HTTP"`
	KnativeProxyHTTPS string `env:"DIREKTIV_KNATIVE_PROXY_HTTPS"`

	FunctionsReconcileInterval int `env:"DIREKTIV_FUNCTIONS_RECONCILE_INTERVAL" envDefault:"1"`

	NatsInstalled bool   `env:"DIREKTIV_NATS_INSTALLED"`
	NatsHost      string `env:"DIREKTIV_NATS_HOST"`
	NatsPort      int    `env:"DIREKTIV_NATS_PORT"      envDefault:"4222"`
	NatsTLS       bool   `env:"DIREKTIV_NATS_TLS"       envDefault:"false"`
	NatsUsername  string `env:"DIREKTIV_NATS_USERNAME"`
	NatsPassword  string `env:"DIREKTIV_NATS_PASSWORD"`

	OidcIssuerUrl string `env:"DIREKTIV_OIDC_ISSUER_URL"`
	OidcClientID  string `env:"DIREKTIV_OIDC_CLIENT_ID"`
	LogsBackend   string `env:"DIREKTIV_LOGS_BACKEND"`
	OtelBackend   string `env:"DIREKTIV_OTEL_BACKEND"`
}

func (conf *Config) GetFunctionsTimeout() time.Duration {
	return time.Second * time.Duration(conf.FunctionsTimeout)
}

func (conf *Config) GetFunctionsReconcileInterval() time.Duration {
	return time.Second * time.Duration(conf.FunctionsReconcileInterval)
}

func (conf *Config) Init() error {
	// Repeat SecretKey length to 16 chars.
	conf.SecretKey += "1234567890123456"
	conf.SecretKey = conf.SecretKey[0:16]

	err := conf.checkInvalidEmptyFields()
	if err != nil {
		return err
	}

	return nil
}

func (conf *Config) checkInvalidEmptyFields() error {
	var invalidEmptyFields []string

	if conf.KnativeServiceAccount == "" {
		invalidEmptyFields = append(invalidEmptyFields, "DIREKTIV_KNATIVE_SERVICE_ACCOUNT")
	}
	if conf.KnativeNamespace == "" {
		invalidEmptyFields = append(invalidEmptyFields, "DIREKTIV_KNATIVE_NAMESPACE")
	}
	if conf.KnativeIngressClass == "" {
		invalidEmptyFields = append(invalidEmptyFields, "DIREKTIV_KNATIVE_INGRESS_CLASS")
	}
	if conf.KnativeSidecar == "" {
		invalidEmptyFields = append(invalidEmptyFields, "DIREKTIV_KNATIVE_SIDECAR")
	}

	if len(invalidEmptyFields) == 0 {
		return nil
	}

	return fmt.Errorf("following fields are required but got empty strings: %v", invalidEmptyFields)
}

type Version struct {
	UnixTime int64 `json:"unix_time"`
}

// TODO: Fix this temp type.
// SyncNamespace is a temporary type until we fix mirroring logic.
type SyncNamespace func(namespace any, mirrorConfig any) (any, error)

type App struct {
	Version *Version
	Config  *Config

	ServiceManager  ServiceManager
	RegistryManager RegistryManager
	GatewayManager  GatewayManager
	SyncNamespace   SyncNamespace
}
