package plugins

import (
	"context"
	"fmt"
	"net/http"
)

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

// type Serve func(w http.ResponseWriter, r *http.Request) bool

func AddPluginToRegistry(plugin Plugin) {
	registry[plugin.Name()] = plugin
}

func GetPluginFromRegistry(plugin string) (Plugin, error) {
	p, ok := registry[plugin]
	if !ok {
		return nil, fmt.Errorf("plugin %s does not exist", plugin)
	}

	return p, nil
}
