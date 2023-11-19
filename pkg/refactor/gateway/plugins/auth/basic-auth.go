package auth

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/consumer"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
)

const (
	BasicAuthPluginName = "basic-auth"
)

// BasicAuthConfig configures a basic-auth plugin instance.
// The plugin can be configured to set consumer information (name, groups, tags).
type BasicAuthConfig struct {
	AddUsernameHeader bool `yaml:"add_username_header"`
	AddTagsHeader     bool `yaml:"add_tags_header"`
	AddGroupsHeader   bool `yaml:"add_groups_header"`
}

type BasicAuthPlugin struct {
	config *BasicAuthConfig
}

func (ba BasicAuthPlugin) Configure(config interface{}) (plugins.Plugin, error) {
	var ok bool
	authConfig := &BasicAuthConfig{}

	if config != nil {
		authConfig, ok = config.(*BasicAuthConfig)
		if !ok {
			return nil, fmt.Errorf("configuration for basic-auth invalid")
		}
	}

	return &BasicAuthPlugin{
		config: authConfig,
	}, nil
}

func (ba BasicAuthPlugin) Name() string {
	return BasicAuthPluginName
}

func (ba BasicAuthPlugin) Type() plugins.PluginType {
	return plugins.AuthPluginType
}

func (ba BasicAuthPlugin) ExecutePlugin(_ context.Context, c *core.Consumer,
	_ http.ResponseWriter, r *http.Request) bool {
	user, pwd, ok := r.BasicAuth()

	// no basic auth provided
	if !ok {
		return true
	}

	consumer := consumer.FindByUser(user)

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

//nolint:gochecknoinits
func init() {
	plugins.AddPluginToRegistry(BasicAuthPlugin{})
}
