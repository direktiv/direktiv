package gateway

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/mitchellh/mapstructure"
)

var registry = make(map[string]core.Plugin)

// RegisterPlugin used to register a new plugin, typically by init() functions.
func RegisterPlugin(p core.Plugin) {
	if os.Getenv("DIREKTIV_APP") != "sidecar" &&
		os.Getenv("DIREKTIV_APP") != "init" {
		slog.Info("adding plugin", slog.String("name", p.Type()))
		registry[p.Type()] = p
	}
}

// NewPlugin creates a new plugin instance from a plugin configuration.
func NewPlugin(config core.PluginConfig) (core.Plugin, error) {
	f, ok := registry[config.Typ]
	if !ok {
		return nil, fmt.Errorf("doesn't exist")
	}

	return f.NewInstance(config)
}

// ConvertConfig only decorates mapstructure.Decode.
func ConvertConfig(config map[string]any, target core.Plugin) error {
	err := mapstructure.Decode(config, target)
	if err != nil {
		return fmt.Errorf("plugin: %s, could not decode plugin config: %w", target.Type(), err)
	}

	return nil
}

// IsJSON helper function checks if a string is a json string.
func IsJSON(str string) bool {
	return json.Unmarshal([]byte(str), &json.RawMessage{}) == nil
}

// WriteJSONError writes error gateway response.
func WriteJSONError(w http.ResponseWriter, status int, endpointFile string, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	inner := struct {
		EndpointFile string `json:"endpointFile,omitempty"`
		Message      any    `json:"message"`
	}{
		EndpointFile: endpointFile,
		Message:      msg,
	}
	payload := struct {
		Error any `json:"error"`
	}{
		Error: inner,
	}

	_ = json.NewEncoder(w).Encode(payload)
}

// WriteInternalError writes error gateway response.
func WriteInternalError(r *http.Request, w http.ResponseWriter, err error, msg string) {
	slog.With("component", "gateway").
		Error(msg, "err", err)
	WriteJSONError(w, http.StatusInternalServerError, ExtractContextEndpoint(r).FilePath, msg)
}

// WriteForbiddenError writes error gateway response.
func WriteForbiddenError(r *http.Request, w http.ResponseWriter, err error, msg string) {
	slog.With("component", "gateway").
		Error(msg, "err", err)
	WriteJSONError(w, http.StatusForbidden, ExtractContextEndpoint(r).FilePath, msg)
}
