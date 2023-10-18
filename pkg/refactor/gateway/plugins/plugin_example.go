package plugins

import (
	"net/http"

	"golang.org/x/exp/slog"
)

// example represents a generic example plugin.
type example struct {
	conf interface{}
}

// buildPlugin initializes the example plugin with the provided configuration and callbacks.
func (e *example) buildPlugin(conf interface{}) (Execute, error) {
	e.conf = conf

	return e.safeProcess, nil
}

// safeProcess is the main execution method of the example plugin.
// It slogs messages and, if a next plugin exists, forwards the request to it.
func (e *example) safeProcess(_ http.ResponseWriter, _ *http.Request) Result {
	slog.Debug("Executed")
	// slog.Debug(e.conf["message"])

	return Result{Status: http.StatusOK}
}

//nolint:gochecknoinits
func init() {
	register(formPluginKey("v1", "example"), &example{})
}

type exampleSpec struct {
	Conf map[string]string `json:"conf"`
}

func (e *example) GetConfigStruct() interface{} {
	return exampleSpec{}
}
