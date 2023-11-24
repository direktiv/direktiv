package auth

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/gateway/consumer"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/direktiv/direktiv/pkg/refactor/spec"
)

const (
	KeyAuthPluginName = "key-auth"
	KeyName           = "API-Token"
)

// KeyAuthConfig configures a key-auth plugin instance.
// The plugin can be configured to set consumer information (name, groups, tags)
// and the name of the header for the api key.
type KeyAuthConfig struct {
	AddUsernameHeader bool `yaml:"add_username_header" mapstructure:"add_username_header"`
	AddTagsHeader     bool `yaml:"add_tags_header" mapstructure:"add_tags_header"`
	AddGroupsHeader   bool `yaml:"add_groups_header" mapstructure:"add_groups_header"`

	// KeyName defines the header for the key
	KeyName string `yaml:"key_name" mapstructure:"key_name"`
}

type KeyAuthPlugin struct {
	config *KeyAuthConfig
}

func ConfigureKeyAuthPlugin(config interface{}) (plugins.PluginInstance, error) {
	keyAuthConfig := &KeyAuthConfig{
		KeyName: KeyName,
	}

	err := plugins.ConvertConfig(BasicAuthPluginName, config, keyAuthConfig)
	if err != nil {
		return nil, err
	}

	return &KeyAuthPlugin{
		config: keyAuthConfig,
	}, nil
}

func (ka *KeyAuthPlugin) ExecutePlugin(ctx context.Context, c *spec.ConsumerFile,
	w http.ResponseWriter, r *http.Request) bool {
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

func (ka *KeyAuthPlugin) Config() interface{} {
	return ka.config
}

//nolint:gochecknoinits
func init() {
	plugins.AddPluginToRegistry(plugins.NewPluginBase(
		KeyAuthPluginName,
		plugins.AuthPluginType,
		ConfigureKeyAuthPlugin))
}
