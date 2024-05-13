package auth

import (
	"fmt"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2"
	"log/slog"
	"net/http"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2/consumer"
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

func (ka *KeyAuthPlugin) ExecutePlugin(c *core.ConsumerFile,
	w http.ResponseWriter, r *http.Request,
) bool {
	key := r.Header.Get(ka.config.KeyName)

	// no basic auth provided
	if key == "" {
		return true
	}

	gwObj := r.Context().Value(plugins.ConsumersParamCtxKey)
	if gwObj == nil {
		slog.Debug("no consumer list in context")

		return true
	}

	consumerList, ok := gwObj.(*consumer.List)
	if !ok {
		plugins.ReportError(r.Context(), w, http.StatusInternalServerError,
			"consumerlist", fmt.Errorf("wrong object in context"))

		return false
	}
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

func (ka *KeyAuthPlugin) Type() string {
	return KeyAuthPluginName
}

//nolint:gochecknoinits
func init() {
	plugins.RegisterPlugin(KeyAuthPluginName, NewKeyAuthPlugin)
}
