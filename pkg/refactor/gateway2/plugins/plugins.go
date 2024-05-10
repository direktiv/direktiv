package plugins

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/direktiv/direktiv/pkg/refactor/core"
)

type Factory func(config core.PluginConfigV2) (core.PluginV2, error)

var registry = make(map[string]Factory)

func RegisterPlugin(name string, factory Factory) {
	if os.Getenv("DIREKTIV_APP") != "sidecar" &&
		os.Getenv("DIREKTIV_APP") != "init" {
		slog.Info("adding plugin", slog.String("name", name))
		registry[name] = factory
	}
}

func NewPlugin(config core.PluginConfigV2) (core.PluginV2, error) {
	f, ok := registry[config.Typ]
	if !ok {
		return nil, fmt.Errorf("unknow plugin '%s'", config.Typ)
	}

	return f(config)
}

func WriteJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	payLoad := struct {
		Data any `json:"data"`
	}{
		Data: v,
	}
	_ = json.NewEncoder(w).Encode(payLoad)
}
