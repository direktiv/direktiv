package plugins

import (
	"fmt"
	"net/http"

	"golang.org/x/exp/slog"
)

var registry = make(map[string]Template)

// this is the contract to provide a Template for plugin.
// NOTE: all Template's must be registred via a init() func using the func register(key string, c Template).
type Template interface {
	// BuildPlugin must instantiate a Plugin and return a function for processing requests or error
	BuildPlugin(conf map[string]string) (Execute, error)
}

// Execute is the final object of a plugin that will be used to process requests.
type Execute func(http.ResponseWriter, *http.Request) Result

type Route struct {
	RouteConfiguration
	plugins []Execute
}

type RouteConfiguration struct {
	Path           string
	Method         string
	Targets        Targets
	TimeoutSeconds int
	PluginsConfig  []Configuration
}

func (conf *Route) Build() error {
	for _, c := range conf.PluginsConfig {
		exc, err := c.Build()
		if err != nil {
			return err
		}
		conf.plugins = append(conf.plugins, exc)
	}

	return nil
}

func (conf *Route) Execute(w http.ResponseWriter, r *http.Request) {
	for _, p := range conf.plugins {
		if resp := p(w, r); resp.Status != http.StatusOK {
			return
		}
	}
}

func (p Configuration) Build() (Execute, error) {
	pt, err := getTemplate(FormPluginKey(p.Version, p.Name))
	if err != nil {
		return nil, err
	}

	return pt.BuildPlugin(p.RuntimeConfig)
}

func FormPluginKey(version, name string) string {
	return version + "/" + name
}

type Target struct {
	Method string
	Host   string
	Path   string
	Scheme string
}

type Targets []Target

type Result struct {
	Status   int    `json:"status"`
	ErrorMsg string `json:"error_msg"`
}

type Configuration struct {
	Name                    string
	Version                 string
	Comment                 string
	Type                    string
	Priority                int               `yaml:"priority"`
	ExecutionTimeoutSeconds int               `yaml:"execution_timeout_seconds"`
	RuntimeConfig           map[string]string `yaml:"runtime_config"`
}

// GetComponent returns the Template specified by name from `Registry`.
func getTemplate(key string) (Template, error) {
	// check if exists
	if _, ok := registry[key]; ok {
		return registry[key], nil
	}

	return nil, fmt.Errorf("%s is not a registered Plugin type", key)
}

// register is called by the `init` function of every `Template`
// this is ment to be used to make a `Template` for a plugin known.
func register(key string, c Template) {
	if _, ok := registry[key]; ok {
		slog.Error("has already been added to the registry", "plugins.Template", key)
	}
	registry[key] = c
}
