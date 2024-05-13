package auth

import (
	"context"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2"
	"log/slog"
	"net/http"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2/plugins"
)

const (
	KeyAuthPluginName = "key-auth"
	KeyName           = "API-Token"
)

// KeyAuthConfig configures a key-auth plugin instance.
// The plugin can be configured to set consumer information (name, groups, tags)
// and the name of the header for the api key.
type KeyAuthConfig struct {
	AddUsernameHeader bool `mapstructure:"add_username_header" yaml:"add_username_header"`
	AddTagsHeader     bool `mapstructure:"add_tags_header"     yaml:"add_tags_header"`
	AddGroupsHeader   bool `mapstructure:"add_groups_header"   yaml:"add_groups_header"`

	// KeyName defines the header for the key
	KeyName string `mapstructure:"key_name" yaml:"key_name"`
}

type KeyAuthPlugin struct {
	config *KeyAuthConfig
}

func NewKeyAuthPlugin(config core.PluginConfigV2) (core.PluginV2, error) {
	keyAuthConfig := &KeyAuthConfig{
		KeyName: KeyName,
	}

	err := gateway2.ConvertConfig(config, keyAuthConfig)
	if err != nil {
		return nil, err
	}

	return &KeyAuthPlugin{
		config: keyAuthConfig,
	}, nil
}

func (ka *KeyAuthPlugin) Execute(w http.ResponseWriter, r *http.Request) (*http.Request, error) {
	// check request is already authenticated
	if gateway2.ReadActiveConsumerFromContext(r) != nil {
		return r, nil
	}

	key := r.Header.Get(ka.config.KeyName)
	// no basic auth provided
	if key == "" {
		return r, nil
	}

	consumerList := gateway2.ReadConsumersListFromContext(r)
	if len(consumerList) == 0 {
		slog.Debug("no consumer configured for api key")

		return r, nil
	}
	c := core.FindConsumerByApiKey(key, consumerList)

	// no consumer with that name
	if c == nil {
		slog.Debug("no consumer configured for api key")

		return r, nil
	}

	if c.APIKey == key {
		// set active comsumer.
		r = r.WithContext(context.WithValue(r.Context(), core.GatewayCtxKeyActiveConsumer, c))

		// set headers if configured.
		if ka.config.AddUsernameHeader {
			r.Header.Set(gateway2.ConsumerUserHeader, c.Username)
		}

		if ka.config.AddTagsHeader && len(c.Tags) > 0 {
			r.Header.Set(gateway2.ConsumerTagsHeader, strings.Join(c.Tags, ","))
		}

		if ka.config.AddGroupsHeader && len(c.Groups) > 0 {
			r.Header.Set(gateway2.ConsumerGroupsHeader, strings.Join(c.Groups, ","))
		}
	}

	return r, nil
}

func (ka *KeyAuthPlugin) Config() interface{} {
	return ka.config
}

func (ka *KeyAuthPlugin) Type() string {
	return KeyAuthPluginName
}

//nolint:gochecknoinits
func init() {
	plugins.RegisterPlugin(KeyAuthPluginName, NewKeyAuthPlugin)
}
