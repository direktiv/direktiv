package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"

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

func (ba BasicAuthPlugin) ExecutePlugin(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {

	fmt.Println("EXECUTE!!!!!!!!!!!!")

	b, err := httputil.DumpRequest(r, false)

	fmt.Println(string(b))
	fmt.Println(err)

	username, password, ok := r.BasicAuth()

	// no basic auth provided
	if !ok {
		return true
	}

	consumer := consumer.ConsumerByUser(username)

	// no consumer with that name
	if consumer == nil {
		slog.Debug("no consumer configured",
			slog.String("user", username))
		return true
	}

	// COMPARE PASSWORD

	// basic auth always returns true to execute other auth plugins
	return true
}

//nolint:gochecknoinits
func init() {
	plugins.AddPluginToRegistry(BasicAuthPlugin{})
}
