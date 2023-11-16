package plugins

import (
	"context"
	"fmt"
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/core"
)

type AuthCtxKey int

const authCtxKey AuthCtxKey = 1

var registry = make(map[string]Plugin)

type PluginType string

var (
	AuthPluginType     PluginType = "auth"
	InboundPluginType  PluginType = "inbound"
	OutboundPluginType PluginType = "outbound"
)

type Plugin interface {
	Configure(config map[string]interface{}) (Plugin, error)
	Name() string
	Type() PluginType
	ExecutePlugin(ctx context.Context, w http.ResponseWriter, r *http.Request) bool
	// getSchema() interface{}
}

func AddPluginToRegistry(plugin Plugin) {
	registry[plugin.Name()] = plugin
}

func AddAuthToContext(ctx context.Context, c *core.Consumer) context.Context {
	return context.WithValue(ctx, authCtxKey, c)
}

func GetPluginFromRegistry(plugin string) (Plugin, error) {
	p, ok := registry[plugin]
	if !ok {
		return nil, fmt.Errorf("plugin %s does not exist", plugin)
	}

	return p, nil
}
