package auth

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2"
)

const (
	DefaultKeyName = "API-Token"
)

type KeyAuthPlugin struct {
	AddUsernameHeader bool `mapstructure:"add_username_header"`
	AddTagsHeader     bool `mapstructure:"add_tags_header"`
	AddGroupsHeader   bool `mapstructure:"add_groups_header"`

	// KeyName defines the header for the key
	KeyName string `mapstructure:"key_name"`
}

func (ka *KeyAuthPlugin) NewInstance(config core.PluginConfigV2) (core.PluginV2, error) {
	pl := &KeyAuthPlugin{
		KeyName: DefaultKeyName,
	}

	err := gateway2.ConvertConfig(config.Config, pl)
	if err != nil {
		return nil, err
	}

	return pl, nil
}

func (ka *KeyAuthPlugin) Execute(w http.ResponseWriter, r *http.Request) *http.Request {
	// check request is already authenticated
	if gateway2.ExtractContextActiveConsumer(r) != nil {
		return r
	}

	key := r.Header.Get(ka.KeyName)
	// no basic auth provided
	if key == "" {
		return r
	}

	consumerList := gateway2.ExtractContextConsumersList(r)
	if consumerList == nil {
		slog.Debug("no consumer configured for api key")

		return r
	}
	c := gateway2.FindConsumerByAPIKey(consumerList, key)
	// no consumer matching auth name
	if c == nil {
		slog.Debug("no consumer configured for api key")

		return r
	}

	if c.APIKey == key {
		// set active consumer
		r = gateway2.InjectContextActiveConsumer(r, c)

		// set headers if configured
		if ka.AddUsernameHeader {
			r.Header.Set(gateway2.ConsumerUserHeader, c.Username)
		}

		if ka.AddTagsHeader && len(c.Tags) > 0 {
			r.Header.Set(gateway2.ConsumerTagsHeader, strings.Join(c.Tags, ","))
		}

		if ka.AddGroupsHeader && len(c.Groups) > 0 {
			r.Header.Set(gateway2.ConsumerGroupsHeader, strings.Join(c.Groups, ","))
		}
	}

	return r
}

func (ka *KeyAuthPlugin) Type() string {
	return "key-auth"
}

func init() {
	gateway2.RegisterPlugin(&KeyAuthPlugin{})
}
