package auth

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"log/slog"
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/gateway/consumer"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
)

type BasicAuthPlugin struct {
	config map[string]interface{}
}

func (ba BasicAuthPlugin) Configure(config map[string]interface{}) (plugins.Plugin, error) {
	return &BasicAuthPlugin{
		config: config,
	}, nil
}

func (ba BasicAuthPlugin) Name() string {
	return "basic-auth"
}

func (ba BasicAuthPlugin) Type() plugins.PluginType {
	return plugins.AuthPluginType
}

func (ba BasicAuthPlugin) ExecutePlugin(ctx context.Context, _ http.ResponseWriter, r *http.Request) bool {
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
		plugins.AddAuthToContext(ctx, consumer)
	}

	// basic auth always returns true to execute other auth plugins
	return true
}

//nolint:gochecknoinits
func init() {
	plugins.AddPluginToRegistry(BasicAuthPlugin{})
}
