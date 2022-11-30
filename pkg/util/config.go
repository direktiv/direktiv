package util

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

// Config contain direktiv configuration.
type Config struct {
	FunctionsService string `yaml:"functions-service"`

	// FunctionsTimeout : Action timeout in milliseconds
	FunctionsTimeout int64 `yaml:"functions-timeout"`

	FlowService string `yaml:"flow-service"`

	PrometheusBackend    string `yaml:"prometheus-backend"`
	OpenTelemetryBackend string `yaml:"opentelemetry-backend"`

	Eventing bool `yaml:"eventing"`
}

// ReadConfig reads direktiv config file.
func ReadConfig(file string) (*Config, error) {

	c := new(Config)

	/* #nosec */
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}

	err = yaml.Unmarshal(data, c)
	if err != nil {
		return nil, fmt.Errorf("failed to parse configuration file: %w", err)
	}

	return c, nil

}

func (cfg *Config) GetTelemetryBackendAddr() string {

	return cfg.OpenTelemetryBackend

}

func (cfg *Config) GetFunctionsTimeout() time.Duration {
	return time.Second * time.Duration(cfg.FunctionsTimeout)
}
