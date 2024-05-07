package plugins

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/mitchellh/mapstructure"
)

const (
	consumerUserHeader   = "Direktiv-Consumer-User"
	consumerTagsHeader   = "Direktiv-Consumer-Tags"
	consumerGroupsHeader = "Direktiv-Consumer-Groups"
)

func NewPlugin(config core.PluginConfigV2) (core.PluginV2, error) {
	//nolint:gocritic
	switch config.Typ {
	case "basic-auth":
		return NewBasicAuthPlugin(config)
	case "debug-target":
		return NewDebugPlugin(config)
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

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	payLoad := struct {
		Data any `json:"data"`
	}{
		Data: v,
	}
	_ = json.NewEncoder(w).Encode(payLoad)
}
