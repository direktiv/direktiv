package plugins

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/mitchellh/mapstructure"
)

var registry = make(map[string]core.PluginV2)

func RegisterPlugin(p core.PluginV2) {
	if os.Getenv("DIREKTIV_APP") != "sidecar" &&
		os.Getenv("DIREKTIV_APP") != "init" {
		slog.Info("adding plugin", slog.String("name", p.Type()))
		registry[p.Type()] = p
	}
}

func NewPlugin(config core.PluginConfigV2) (core.PluginV2, error) {
	f, ok := registry[config.Typ]
	if !ok {
		return nil, fmt.Errorf("unknow plugin '%s'", config.Typ)
	}

	return f.NewInstance(config)
}

func ConvertConfig(config map[string]any, target any) error {
	err := mapstructure.Decode(config, target)
	if err != nil {
		return errors.Join(err, errors.New("configuration invalid"))
	}

	return nil
}

func IsJSON(str string) bool {
	var js json.RawMessage

	return json.Unmarshal([]byte(str), &js) == nil
}
