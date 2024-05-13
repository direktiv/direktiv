package plugins

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/mitchellh/mapstructure"
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

func ConvertConfig(config any, target any) error {
	err := mapstructure.Decode(config, target)
	if err != nil {
		return errors.Join(err, errors.New("configuration invalid"))
	}

	return nil
}
