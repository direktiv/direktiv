package plugins

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/core"
)

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

	err := ConvertConfig(config, authConfig)
	if err != nil {
		return nil, err
	}

	return &BasicAuthPlugin{
		config: authConfig,
	}, nil
}

func (ba *BasicAuthPlugin) Execute(w http.ResponseWriter, r *http.Request) (*http.Request, error) {
	user, pwd, ok := r.BasicAuth()

	// no basic auth provided
	if !ok {
		return r, nil
	}

	slog.Debug("running basic-auth plugin", "user", user)

	gwObj := r.Context().Value(core.GatewayCtxKeyConsumers)
	if gwObj == nil {
		slog.Debug("no consumer list in context", slog.String("user", user))

		return r, nil
	}
	consumerList, ok := gwObj.([]core.ConsumerV2)
	if !ok {
		return nil, errors.New("missing consumer list in context")
	}
	consumer := core.FindConsumerByUser(user, consumerList)

	// no consumer with that name
	if consumer == nil {
		slog.Debug("no consumer configured",
			slog.String("user", user))

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
			r.Header.Set(consumerUserHeader, consumer.Username)
		}

		if ba.config.AddTagsHeader && len(consumer.Tags) > 0 {
			r.Header.Set(consumerTagsHeader, strings.Join(consumer.Tags, ","))
		}

		if ba.config.AddGroupsHeader && len(consumer.Groups) > 0 {
			r.Header.Set(consumerGroupsHeader, strings.Join(consumer.Groups, ","))
		}
	}

	return r, nil
}

func (ba *BasicAuthPlugin) Type() string {
	return "basic-auth"
}

func (ba *BasicAuthPlugin) Config() interface{} {
	return ba.config
}

//nolint:gochecknoinits
func init() {
	RegisterPlugin("basic-auth", NewBasicAuthPlugin)
}
