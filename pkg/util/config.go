package util

import (
	"errors"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config contain direktiv configuration
type Config struct {
	FunctionsService string `yaml:"functions-service"`
	FlowService      string `yaml:"flow-service"`

	PrometheusBackend string `yaml:"prometheus-backend"`
	RedisBackend      string `yaml:"redis-backend"`

	OpenTelemetryBackend string `yaml:"opentelemetry-backend"`
}

// ReadConfig reads direktiv config file
func ReadConfig(file string) (*Config, error) {

	c := new(Config)

	/* #nosec */
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, c)
	if err != nil {
		return nil, err
	}

	return c, nil

}

func (cfg *Config) GetTelemetryBackendAddr() string {

	fmt.Println(cfg.OpenTelemetryBackend)

	if cfg.OpenTelemetryBackend == "" {
		panic(errors.New("need to configure telemetry"))
	}

	if cfg.OpenTelemetryBackend == "none" {
		return ""
	}

	if cfg.OpenTelemetryBackend == "host" {
		/*
			s := os.Getenv("DIREKTIV_NODE_IP")
			if s == "" {
				panic(errors.New("need to configure telemetry environment variable DIREKTIV_NODE_IP"))
			}
			return s + ":4317"
		*/
		return "direktiv-otel-collector.default:4317"
	}

	return cfg.OpenTelemetryBackend

}
