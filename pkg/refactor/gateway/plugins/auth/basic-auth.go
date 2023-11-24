package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"log/slog"
	"net/http"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/gateway/consumer"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/direktiv/direktiv/pkg/refactor/spec"
)

const (
	BasicAuthPluginName = "basic-auth"
)

// BasicAuthConfig configures a basic-auth plugin instance.
// The plugin can be configured to set consumer information (name, groups, tags).
type BasicAuthConfig struct {
	AddUsernameHeader bool `yaml:"add_username_header" mapstructure:"add_username_header"`
	AddTagsHeader     bool `yaml:"add_tags_header" mapstructure:"add_tags_header"`
	AddGroupsHeader   bool `yaml:"add_groups_header" mapstructure:"add_groups_header"`
}

type BasicAuthPlugin struct {
	config *BasicAuthConfig
}

func ConfigureBasicAuthPlugin(config interface{}, ns string) (plugins.PluginInstance, error) {
	authConfig := &BasicAuthConfig{}

	err := plugins.ConvertConfig(BasicAuthPluginName, config, &authConfig)
	if err != nil {
		return nil, err
	}

	return &BasicAuthPlugin{
		config: authConfig,
	}, nil
}

func (ba *BasicAuthPlugin) ExecutePlugin(c *spec.ConsumerFile,
	w http.ResponseWriter, r *http.Request) bool {

	user, pwd, ok := r.BasicAuth()

	// no basic auth provided
	if !ok {
		return true
	}

	gwObj := r.Context().Value(plugins.ConsumersParamCtxKey)
	if gwObj == nil {
		slog.Debug("no consumer list in context",
			slog.String("user", user))

		return true
	}
	consumerList := gwObj.(*consumer.ConsumerList)
	consumer := consumerList.FindByUser(user)

	// no consumer with that name
	if consumer == nil {
		slog.Debug("no consumer configured",
			slog.String("user", user))

		return true
	}

	// comparing passwords
	userHash := sha256.Sum256([]byte(user))
	pwdHash := sha256.Sum256([]byte(pwd))
	userHashExpected := sha256.Sum256([]byte(consumer.Username))
	pwdHashExpected := sha256.Sum256([]byte(consumer.Password))

	usernameMatch := (subtle.ConstantTimeCompare(userHash[:], userHashExpected[:]) == 1)
	passwordMatch := (subtle.ConstantTimeCompare(pwdHash[:], pwdHashExpected[:]) == 1)

	if usernameMatch && passwordMatch {
		*c = *consumer

		// set headers if configured.
		if ba.config.AddUsernameHeader {
			r.Header.Set(plugins.ConsumerUserHeader, c.Username)
		}

		if ba.config.AddTagsHeader && len(c.Tags) > 0 {
			r.Header.Set(plugins.ConsumerTagsHeader, strings.Join(c.Tags, ","))
		}

		if ba.config.AddGroupsHeader && len(c.Groups) > 0 {
			r.Header.Set(plugins.ConsumerGroupsHeader, strings.Join(c.Groups, ","))
		}
	}

	// basic auth always returns true to execute other auth plugins
	return true

}

func (ba *BasicAuthPlugin) Config() interface{} {
	return ba.config
}

//nolint:gochecknoinits
func init() {
	plugins.AddPluginToRegistry(plugins.NewPluginBase(
		BasicAuthPluginName,
		plugins.AuthPluginType,
		ConfigureBasicAuthPlugin))
}
