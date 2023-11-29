package endpoints

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
)

type EndpointList struct {
	currentTree *node

	lock sync.Mutex
}

type Endpoint struct {
	EndpointBase            *core.EndpointBase
	Namespace               string
	FilePath                string
	AuthPluginInstances     []plugins.PluginInstance
	InboundPluginInstances  []plugins.PluginInstance
	TargetPluginInstance    plugins.PluginInstance
	OutboundPluginInstances []plugins.PluginInstance
	Errors                  []string
	Warnings                []string
}

func NewEndpointList() *EndpointList {
	return &EndpointList{
		currentTree: &node{},
	}
}

func (e *EndpointList) Routes() []Route {
	return e.currentTree.Routes()
}

func (e *EndpointList) FindRoute(route, method string) (*Endpoint, map[string]string) {
	if !strings.HasPrefix(route, "/") {
		route = "/" + route
	}

	// if unknown method
	m, ok := methodMap[method]
	if !ok {
		return nil, nil
	}

	routeCtx := NewRouteContext()
	_, _, endpointEntry := e.currentTree.FindRoute(routeCtx, m, route)
	if endpointEntry == nil {
		return nil, nil
	}

	// add path extension variables in context, e.g. /{id}
	urlParams := make(map[string]string)
	for i := 0; i < len(routeCtx.URLParams.Keys); i++ {
		key := routeCtx.URLParams.Keys[i]
		urlParams[key] = routeCtx.URLParams.Values[i]
	}

	return endpointEntry, urlParams
}

func (e *EndpointList) GetEndpoints() []core.EndpointListItem {
	routes := e.Routes()
	items := make([]core.EndpointListItem, 0)

	for i := range routes {
		r := routes[i]
		methods := make([]string, 0)
		ep := core.EndpointListItem{
			Path:    r.FilePath,
			Pattern: r.Pattern,
		}
		for m, h := range r.Handlers {
			methods = append(methods, m)
			ep.Warnings = h.Warnings
			ep.Errors = h.Errors
			ep.AllowAnonymous = h.EndpointBase.AllowAnonymous
			ep.PathExtension = h.EndpointBase.PathExtension
			ep.Plugins = h.EndpointBase.Plugins
			ep.Timeout = h.EndpointBase.Timeout
		}
		ep.Methods = methods
		items = append(items, ep)
	}

	return items
}

func (e *EndpointList) SetEndpoints(endpointList []*Endpoint) {
	newTree := &node{}

	for i := range endpointList {
		ep := endpointList[i]

		// skip the files with invalid content
		if ep.EndpointBase == nil {
			continue
		}

		slog.Debug("adding endpoint",
			slog.String("path", ep.FilePath),
			slog.String("extension", ep.EndpointBase.PathExtension))

		// remove the file extension, most likely .yaml
		storePath := strings.TrimSuffix(ep.FilePath, filepath.Ext(ep.FilePath))

		// add path extension if there is any
		if ep.EndpointBase.PathExtension != "" {
			storePath = filepath.Join(storePath, ep.EndpointBase.PathExtension)
		}

		buildPluginChain(ep)

		// assign handler to all methods
		for a := range ep.EndpointBase.Methods {
			m := ep.EndpointBase.Methods[a]
			mMethod, ok := methodMap[m]
			if !ok {
				slog.Warn("http method unknown",
					slog.String("endpoint", ep.FilePath),
					slog.String("method", m))

				continue
			}

			slog.Info("adding endpoint",
				slog.String("path", storePath),
				slog.String("method", m))

			newTree.InsertRoute(mMethod, storePath, ep)
		}
	}

	// replace real tree with new one
	e.lock.Lock()
	defer e.lock.Unlock()
	e.currentTree = newTree
}

func buildPluginChain(endpoint *Endpoint) {
	slog.Info("building plugin chain for endpoint",
		slog.String("endpoint", endpoint.FilePath))

	// add target if set
	if endpoint.EndpointBase.Plugins.Target.Type != "" {
		targetPlugin, err := configurePlugin(endpoint.EndpointBase.Plugins.Target,
			plugins.TargetPluginType, endpoint.Namespace)
		if err != nil {
			endpoint.Errors = append(endpoint.Errors, err.Error())
		} else {
			endpoint.TargetPluginInstance = targetPlugin
		}
	} else {
		endpoint.Warnings = append(endpoint.Warnings, "no target plugin set")
	}

	// add auth plugins
	authPlugins, errors := processPlugins(endpoint.EndpointBase.Plugins.Auth,
		plugins.AuthPluginType, endpoint.Namespace)
	endpoint.AuthPluginInstances = authPlugins
	endpoint.Errors = append(endpoint.Errors, errors...)

	// inbound
	inboundPlugins, errors := processPlugins(endpoint.EndpointBase.Plugins.Inbound,
		plugins.InboundPluginType, endpoint.Namespace)
	endpoint.InboundPluginInstances = inboundPlugins
	endpoint.Errors = append(endpoint.Errors, errors...)

	// outbound
	outboundPlugins, errors := processPlugins(endpoint.EndpointBase.Plugins.Outbound,
		plugins.OutboundPluginType, endpoint.Namespace)
	endpoint.OutboundPluginInstances = outboundPlugins
	endpoint.Errors = append(endpoint.Errors, errors...)
}

func processPlugins(pluginConfigs []core.PluginConfig, t plugins.PluginType, ns string) ([]plugins.PluginInstance, []string) {
	errors := make([]string, 0)
	configuredPlugins := make([]plugins.PluginInstance, 0)

	for i := range pluginConfigs {
		config := pluginConfigs[i]
		pi, err := configurePlugin(config, t, ns)
		if err != nil {
			// add error of the plugin to error array
			errors = append(errors, fmt.Sprintf("%s: %s", config.Type, err.Error()))

			continue
		}
		configuredPlugins = append(configuredPlugins, pi)
	}

	return configuredPlugins, errors
}

func configurePlugin(config core.PluginConfig, t plugins.PluginType, ns string) (plugins.PluginInstance, error) {
	slog.Info("processing plugin",
		slog.String("plugin", config.Type))

	p, err := plugins.GetPluginFromRegistry(config.Type)
	if err != nil {
		return nil, err
	}

	if p.Type() != t {
		slog.Error("plugin type mismatch", slog.String(string(p.Type()), string(t)))

		return nil, fmt.Errorf("plugin %s not of type %s", config.Type, t)
	}

	return p.Configure(config.Configuration, ns)
}
