package plugins

import (
	"fmt"
	"net/http"

	"golang.org/x/exp/slog"
)

var registry = make(map[string]template)

// this is the contract to provide a template for plugin.
// NOTE: all template's must be registred via a init() func using the func register(key string, c template).
type template interface {
	// buildPlugin must instantiate a Plugin and return a function for processing requests or error
	buildPlugin(conf interface{}) (Execute, error)
}

// Execute is the final object of a plugin that will be used to process requests.
type Execute func(http.ResponseWriter, *http.Request) Result

func (p Configuration) Build() (Execute, error) {
	pt, err := getTemplate(formPluginKey(p.Version, p.Name))
	if err != nil {
		return nil, err
	}

	return pt.buildPlugin(p.RuntimeConfig)
}

func formPluginKey(version, name string) string {
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
	Name          string
	Version       string
	RuntimeConfig interface{} `yaml:"runtime_config"`
}

// GetComponent returns the Template specified by name from `Registry`.
func getTemplate(key string) (template, error) {
	// check if exists
	if _, ok := registry[key]; ok {
		return registry[key], nil
	}

	return nil, fmt.Errorf("%s is not a registered Plugin type", key)
}

// register is called by the `init` function of every `Template`
// this is ment to be used to make a `Template` for a plugin known.
func register(key string, c template) {
	if _, ok := registry[key]; ok {
		slog.Error("has already been added to the registry", "plugins.Template", key)
	}
	registry[key] = c
}
