package auth

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2/plugins"
)

const (
	keyAuthPluginName = "key-auth"
	keyName           = "API-Token"
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
		KeyName: keyName,
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
	if gateway2.ParseRequestActiveConsumer(r) != nil {
		return r, nil
	}

	key := r.Header.Get(ka.config.KeyName)
	// no basic auth provided
	if key == "" {
		return r, nil
	}

	consumerList := gateway2.ParseRequestConsumersList(r)
	if consumerList == nil {
		slog.Debug("no consumer configured for api key")

		return r, nil
	}
	c := gateway2.FindConsumerByAPIKey(consumerList, key)
	// no consumer matching auth name
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
	return keyAuthPluginName
}

//nolint:gochecknoinits
func init() {
	plugins.RegisterPlugin(keyAuthPluginName, NewKeyAuthPlugin)
}
