package plugins

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/core"
)

// nolint
const (
	ConsumerUserHeader   = "Direktiv-Consumer-User"
	ConsumerTagsHeader   = "Direktiv-Consumer-Tags"
	ConsumerGroupsHeader = "Direktiv-Consumer-Groups"
)

var registry = make(map[string]Plugin)

type PluginType string

var (
	AuthPluginType     PluginType = "auth"
	InboundPluginType  PluginType = "inbound"
	OutboundPluginType PluginType = "outbound"
)

type Plugin interface {
	Configure(config interface{}) (Plugin, error)
	Name() string
	Type() PluginType
	ExecutePlugin(ctx context.Context, c *core.Consumer,
		w http.ResponseWriter, r *http.Request) bool
}

func AddPluginToRegistry(plugin Plugin) {
	slog.Info("adding plugin", slog.String("name", plugin.Name()))
	registry[plugin.Name()] = plugin
}

func GetPluginFromRegistry(plugin string) (Plugin, error) {
	p, ok := registry[plugin]
	if !ok {
		return nil, fmt.Errorf("plugin %s does not exist", plugin)
	}

	return p, nil
}

var URLParamCtxKey = &ContextKey{"URLParamContext"}

type ContextKey struct {
	name string
}

func (k *ContextKey) String() string {
	return "plugin context value " + k.name
}

func ReportError(w http.ResponseWriter, status int, msg string, err error) {
	slog.Error("can not process plugin", slog.String("error", err.Error()))
	w.WriteHeader(status)
	errMsg := fmt.Sprintf("%s: %s", msg, err.Error())

	// nolint
	w.Write([]byte(errMsg))
}
