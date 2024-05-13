package auth

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2/plugins"
)

const basicAuthPluginName = "basic-auth"

// BasicAuthConfig configures a basic-auth plugin instance.
// The plugin can be configured to set consumer information (name, groups, tags).
type BasicAuthConfig struct {
	AddUsernameHeader bool `mapstructure:"add_username_header" yaml:"add_username_header"`
	AddTagsHeader     bool `mapstructure:"add_tags_header"     yaml:"add_tags_header"`
	AddGroupsHeader   bool `mapstructure:"add_groups_header"   yaml:"add_groups_header"`
}

type BasicAuthPlugin struct {
	config *BasicAuthConfig
}

var _ core.PluginV2 = &BasicAuthPlugin{}

func NewBasicAuthPlugin(config core.PluginConfigV2) (core.PluginV2, error) {
	authConfig := &BasicAuthConfig{}

	err := gateway2.ConvertConfig(config, authConfig)
	if err != nil {
		return nil, err
	}

	return &BasicAuthPlugin{
		config: authConfig,
	}, nil
}

func (ba *BasicAuthPlugin) Execute(w http.ResponseWriter, r *http.Request) (*http.Request, error) {
	// check request is already authenticated
	if gateway2.ReadActiveConsumerFromContext(r) != nil {
		return r, nil
	}
	user, pwd, ok := r.BasicAuth()
	// no basic auth provided
	if !ok {
		return r, nil
	}

	consumerList := gateway2.ReadConsumersListFromContext(r)
	if consumerList == nil {
		return r, nil
	}
	consumer := gateway2.FindConsumerByUser(consumerList, user)
	// no consumer matching auth name
	if consumer == nil {
		return r, nil
	}

	// comparing passwords
	userHash := sha256.Sum256([]byte(user))
	pwdHash := sha256.Sum256([]byte(pwd))
	userHashExpected := sha256.Sum256([]byte(consumer.Username))
	pwdHashExpected := sha256.Sum256([]byte(consumer.Password))

	usernameMatch := subtle.ConstantTimeCompare(userHash[:], userHashExpected[:]) == 1
	passwordMatch := subtle.ConstantTimeCompare(pwdHash[:], pwdHashExpected[:]) == 1

	if usernameMatch && passwordMatch {
		// set active comsumer.
		r = r.WithContext(context.WithValue(r.Context(), core.GatewayCtxKeyActiveConsumer, consumer))
		// set headers if configured.
		if ba.config.AddUsernameHeader {
			r.Header.Set(gateway2.ConsumerUserHeader, consumer.Username)
		}

		if ba.config.AddTagsHeader && len(consumer.Tags) > 0 {
			r.Header.Set(gateway2.ConsumerTagsHeader, strings.Join(consumer.Tags, ","))
		}

		if ba.config.AddGroupsHeader && len(consumer.Groups) > 0 {
			r.Header.Set(gateway2.ConsumerGroupsHeader, strings.Join(consumer.Groups, ","))
		}
	}

	return r, nil
}

func (ba *BasicAuthPlugin) Type() string {
	return basicAuthPluginName
}

func (ba *BasicAuthPlugin) Config() interface{} {
	return ba.config
}

//nolint:gochecknoinits
func init() {
	plugins.RegisterPlugin(basicAuthPluginName, NewBasicAuthPlugin)
}
