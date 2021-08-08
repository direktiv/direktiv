package api

import (
	"flag"
	"fmt"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/pkg/util"
	"gopkg.in/yaml.v2"
)

// Templates ..
type Templates struct {
	WorkflowTemplateDirectories []NamedDirectory
	ActionTemplateDirectories   []NamedDirectory
}

// Config ..
type Config struct {
	BlockList string    `yaml:"blocklist"`
	Templates Templates `yaml:"templates"`
}

const apiBind = "0.0.0.0:8080"

func configCheck(c *Config) error {
	if util.IngressEndpoint() == "" {
		return fmt.Errorf("api ingress endpoint not set")
	}
	return nil
}

// ConfigFromFile reads API configuration from file
func ConfigFromFile(cfgPath string) (*Config, error) {

	cfg := new(Config)

	log.Debugf("reading config %s", cfgPath)

	/* #nosec G304 */
	b, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		log.Errorf("can not read config file: %v", err)
		return nil, err
	}

	err = yaml.Unmarshal(b, cfg)
	if err != nil {
		log.Errorf("can not unmarshal config file: %v", err)
		return nil, err
	}

	return cfg, configCheck(cfg)
}

// Configure reads config from file or env
func Configure() (*Config, error) {
	var (
		cfgPath string
		cfg     *Config
		err     error
	)

	flag.StringVar(&cfgPath, "c", "", "points to api server configuration file")
	flag.Parse()

	if cfgPath == "" {
		return cfg, fmt.Errorf("configuration file not provided")
	}

	cfg, err = ConfigFromFile(cfgPath)

	return cfg, err

}

func (c *Config) hasBlockList() bool {
	if c.BlockList == "" {
		return false
	}
	return true
}

func (c *Config) hasWorkflowTemplateDefault() bool {
	if c.Templates.WorkflowTemplateDirectories == nil {
		return false
	}

	for _, tuple := range c.Templates.WorkflowTemplateDirectories {
		if tuple.Label == "default" {
			return true
		}
	}

	return false
}

func (c *Config) hasActionTemplateDefault() bool {
	if c.Templates.ActionTemplateDirectories == nil {
		return false
	}

	for _, tuple := range c.Templates.ActionTemplateDirectories {
		if tuple.Label == "default" {
			return true
		}
	}

	return false
}
