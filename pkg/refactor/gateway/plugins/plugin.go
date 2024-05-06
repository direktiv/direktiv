package plugins

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/direktiv/direktiv/pkg/flow/nohome/recipient"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/mitchellh/mapstructure"
	"go.opentelemetry.io/otel/trace"
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
	TargetPluginType   PluginType = "target"
	InboundPluginType  PluginType = "inbound"
	OutboundPluginType PluginType = "outbound"
)

type Plugin interface {
	Configure(config interface{}, namespace string) (core.PluginInstance, error)
	Name() string
	Type() PluginType
}

// PluginBase is a basic implementation of the Plugin interface.
type PluginBase struct {
	pname    string
	ptype    PluginType
	configFn func(interface{}, string) (core.PluginInstance, error)
}

func (p PluginBase) Name() string {
	return p.pname
}

func (p PluginBase) Type() PluginType {
	return p.ptype
}

func (p PluginBase) Configure(config interface{}, ns string) (core.PluginInstance, error) {
	return p.configFn(config, ns)
}

func NewPluginBase(pname string, ptype PluginType,
	configFn func(interface{}, string) (core.PluginInstance, error),
) Plugin {
	return &PluginBase{
		pname:    pname,
		ptype:    ptype,
		configFn: configFn,
	}
}

// ConvertConfig converts an interface into the config struct of the plugin.
// It is used in the `Configure` function of the Plugin.
func ConvertConfig(config interface{}, target interface{}) error {
	if config != nil {
		err := mapstructure.Decode(config, target)
		if err != nil {
			return errors.Join(err, errors.New("configuration invalid"))
		}
	}

	return nil
}

func GetAllPlugins() map[string]Plugin {
	return registry
}

func AddPluginToRegistry(plugin Plugin) {
	if os.Getenv("DIREKTIV_APP") != "sidecar" &&
		os.Getenv("DIREKTIV_APP") != "init" {
		slog.Info("adding plugin", slog.String("name", plugin.Name()))
		registry[plugin.Name()] = plugin
	}
}

func GetPluginFromRegistry(plugin string) (Plugin, error) {
	p, ok := registry[plugin]
	if !ok {
		return nil, fmt.Errorf("plugin %s does not exist", plugin)
	}

	return p, nil
}

var (
	URLParamCtxKey       = &ContextKey{"URLParamContext"}
	ConsumersParamCtxKey = &ContextKey{"ConsumersParamCtxKey"}
	NamespaceCtxKey      = &ContextKey{"namespace"}
	EndpointCtxKey       = &ContextKey{"endpoint"}
	RouteCtxKey          = &ContextKey{"route"}
)

type ContextKey struct {
	name string
}

func (k *ContextKey) String() string {
	return "plugin context value " + k.name
}

func ReportError(ctx context.Context, w http.ResponseWriter, status int, msg string, err error) {
	span := trace.SpanFromContext(ctx)
	defer span.End()
	spanContext := span.SpanContext()
	traceID := spanContext.TraceID().String()
	spanID := spanContext.SpanID()
	ns, ok := ctx.Value(NamespaceCtxKey).(string)
	if !ok {
		slog.Error("TODO: This must be a bug, fixme A")
	}
	endP, ok := ctx.Value(EndpointCtxKey).(string)
	if !ok {
		slog.Error("TODO: This must be a bug, fixme B")
	}
	routePath, ok := ctx.Value(RouteCtxKey).(string)
	if !ok {
		slog.Error("TODO: This must be a bug, fixme C")
	}
	slog.Error("can not process plugin",
		"namespace", ns,
		"trace", traceID,
		"span", spanID,
		"endpoint", endP,
		"route", routePath,
		"track", recipient.Route.String()+"."+ns+"."+endP,
		"err", err,
	)

	// TODO: Metrics
	w.WriteHeader(status)
	errMsg := fmt.Sprintf("%s: %s", msg, err.Error())

	// nolint
	w.Write([]byte(errMsg))
}

func ReportNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	// nolint
	w.Write([]byte("not found"))
}

func IsJSON(str string) bool {
	var js json.RawMessage

	return json.Unmarshal([]byte(str), &js) == nil
}
