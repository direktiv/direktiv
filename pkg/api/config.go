package api

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml"
)

// Config ..
type Config struct {
	Ingress struct {
		Endpoint string
	}

	Server struct {
		Bind string
	}
}

const (
	direktivAPIBind    = "DIREKTIV_API_BIND"
	direktivAPIIngress = "DIREKTIV_API_INGRESS"
)

func configCheck(c *Config) error {
	if c.Ingress.Endpoint == "" || c.Server.Bind == "" {
		return fmt.Errorf("api bind or ingress endpoint not set")
	}
	return nil
}

// ConfigFromEnv reads API configuration from env variables
func ConfigFromEnv() (*Config, error) {

	c := &Config{}
	c.Ingress.Endpoint = os.Getenv(direktivAPIIngress)
	c.Server.Bind = os.Getenv(direktivAPIBind)

	if c.Ingress.Endpoint == "" || c.Server.Bind == "" {
		return nil, fmt.Errorf("api bind or ingress endpoint not set")
	}

	return c, configCheck(c)
}

// ConfigFromFile reads API configuration from file
func ConfigFromFile(cfgPath string) (*Config, error) {

	cfg := new(Config)
	r, err := os.Open(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %s", err.Error())
	}
	defer r.Close()

	dec := toml.NewDecoder(r)
	err = dec.Decode(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file contents: %s", err.Error())
	}

	return cfg, configCheck(cfg)
}
