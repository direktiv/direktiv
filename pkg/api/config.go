package api

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/pelletier/go-toml"
)

// Config ..
type Config struct {
	Ingress struct {
		Endpoint string
		TLS      bool
	}

	Server struct {
		Bind string
	}

	Templates struct {
		WorkflowTemplateDirectories []NamedDirectory
		ActionTemplateDirectories   []NamedDirectory
	}
}

const (
	direktivAPIBind            = "DIREKTIV_API_BIND"
	direktivAPIIngress         = "DIREKTIV_API_INGRESS"
	direktivWFTemplateDirs     = "DIREKTIV_WF_TEMPLATES"
	direktivActionTemplateDirs = "DIREKTIV_ACTION_TEMPLATES"
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

	x := os.Getenv(direktivWFTemplateDirs)
	if x != "" {
		c.Templates.WorkflowTemplateDirectories = make([]NamedDirectory, 0)
		elems := strings.Split(x, ":")
		for _, t := range elems {
			y := strings.Split(t, "=")
			if len(y) != 2 {
				// invalid string, should be LABEL=PATH
				continue
			}

			c.Templates.WorkflowTemplateDirectories = append(c.Templates.WorkflowTemplateDirectories, NamedDirectory{
				Label:     y[0],
				Directory: y[1],
			})
		}
	}

	x = os.Getenv(direktivActionTemplateDirs)
	if x != "" {
		c.Templates.ActionTemplateDirectories = make([]NamedDirectory, 0)
		elems := strings.Split(x, ":")
		for _, t := range elems {
			y := strings.Split(t, "=")
			if len(y) != 2 {
				// invalid string, should be LABEL=PATH
				continue
			}

			c.Templates.ActionTemplateDirectories = append(c.Templates.ActionTemplateDirectories, NamedDirectory{
				Label:     y[0],
				Directory: y[1],
			})
		}
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
		cfg, err = ConfigFromEnv()
	} else {
		cfg, err = ConfigFromFile(cfgPath)
	}

	return cfg, err

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
