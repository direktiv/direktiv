package auth

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/consumer"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

const (
	KeyAuthPluginName = "key-auth"
	KeyName           = "API-Token"
)

// KeyAuthConfig configures a key-auth plugin instance.
// The plugin can be configured to set consumer information (name, groups, tags)
// and the name of the header for the api key.
type KeyAuthConfig struct {
	AddUsernameHeader bool `yaml:"add_username_header"`
	AddTagsHeader     bool `yaml:"add_tags_header"`
	AddGroupsHeader   bool `yaml:"add_groups_header"`

	// KeyName defines the header for the key
	KeyName string `yaml:"key_name"`
}

type KeyAuthPlugin struct {
	config *KeyAuthConfig
}

func (ka KeyAuthPlugin) Configure(config interface{}) (plugins.PluginInstance, error) {
	keyAuthConfig := &KeyAuthConfig{
		KeyName: KeyName,
	}

	if config != nil {
		err := mapstructure.Decode(config, &keyAuthConfig)
		if err != nil {
			return nil, errors.Wrap(err, "configuration for target-flow invalid")
		}

		if keyAuthConfig.KeyName == "" {
			keyAuthConfig.KeyName = KeyName
		}
	}

	return &KeyAuthPlugin{
		config: keyAuthConfig,
	}, nil
}

func (ka KeyAuthPlugin) Name() string {
	return KeyAuthPluginName
}

func (ka KeyAuthPlugin) Type() plugins.PluginType {
	return plugins.AuthPluginType
}

func (ka KeyAuthPlugin) ExecutePlugin(ctx context.Context, c *core.Consumer,
	_ http.ResponseWriter, r *http.Request) bool {

	key := r.Header.Get(ka.config.KeyName)

	// no basic auth provided
	if key == "" {
		return true
	}

	gwObj := ctx.Value(plugins.ConsumersParamCtxKey)
	if gwObj == nil {
		slog.Debug("no consumer list in context")

		return true
	}
	consumerList := gwObj.(*consumer.ConsumerList)
	consumer := consumerList.FindByAPIKey(key)

	// no consumer with that name
	if consumer == nil {
		slog.Debug("no consumer configured for api key")

		return true
	}

	if consumer.APIKey == key {
		*c = *consumer

		// set headers if configured.
		if ka.config.AddUsernameHeader {
			r.Header.Set(plugins.ConsumerUserHeader, c.Username)
		}

		if ka.config.AddTagsHeader && len(c.Tags) > 0 {
			r.Header.Set(plugins.ConsumerTagsHeader, strings.Join(c.Tags, ","))
		}

		if ka.config.AddGroupsHeader && len(c.Groups) > 0 {
			r.Header.Set(plugins.ConsumerGroupsHeader, strings.Join(c.Groups, ","))
		}
	}

	return true
}

//nolint:gochecknoinits
func init() {
	plugins.AddPluginToRegistry(KeyAuthPlugin{})
}
