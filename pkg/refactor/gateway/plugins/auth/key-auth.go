package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/consumer"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
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

func (ka KeyAuthPlugin) Configure(config interface{}) (plugins.Plugin, error) {
	var ok bool
	keyAuthConfig := &KeyAuthConfig{
		KeyName: KeyName,
	}

	if config != nil {
		keyAuthConfig, ok = config.(*KeyAuthConfig)
		if !ok {
			return nil, fmt.Errorf("configuration for key-auth invalid")
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

func (ka KeyAuthPlugin) ExecutePlugin(_ context.Context, c *core.Consumer,
	_ http.ResponseWriter, r *http.Request) bool {

	key := r.Header.Get(ka.config.KeyName)

	// no basic auth provided
	if key == "" {
		return true
	}

	consumer := consumer.FindByAPIKey(key)

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
