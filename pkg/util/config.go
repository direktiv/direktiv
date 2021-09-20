package util

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config contain direktiv configuration
type Config struct {
	FunctionsService string `yaml:"functions-service"`
	FlowService      string `yaml:"flow-service"`

	PrometheusBackend string `yaml:"prometheus-backend"`
	RedisBackend      string `yaml:"redis-backend"`
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
