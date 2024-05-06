package plugins

import (
	"errors"
	"fmt"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/mitchellh/mapstructure"
)

func NewPlugin(config core.PluginConfigV2) (core.PluginV2, error) {
	//nolint:gocritic
	switch config.Typ {
	case "basic-auth":
		return NewBasicAuthPlugin(config)
	}

	return nil, fmt.Errorf("unknow plugin '%s'", config.Typ)
}

func ConvertConfig(config any, target any) error {
	err := mapstructure.Decode(config, target)
	if err != nil {
		return errors.Join(err, errors.New("configuration invalid"))
	}

	return nil
}
