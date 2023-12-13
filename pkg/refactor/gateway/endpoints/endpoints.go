package endpoints

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
)

type EndpointList struct {
	currentTree *node

	setList []*core.Endpoint

	lock sync.Mutex
}

func NewEndpointList() *EndpointList {
	return &EndpointList{
		currentTree: &node{},
	}
}

func (e *EndpointList) Routes() []Route {
	return e.currentTree.Routes()
}

func (e *EndpointList) FindRoute(route, method string) (*core.Endpoint, map[string]string) {
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

func (e *EndpointList) GetEndpoints() []*core.Endpoint {
	return e.setList
}

func (e *EndpointList) SetEndpoints(endpointList []*core.Endpoint) {
	newTree := &node{}

	for i := range endpointList {
		ep := endpointList[i]

		if ep.Path == "" {
			slog.Warn("no path configured for route", "path", ep.FilePath)

			continue
		}

		if !strings.HasPrefix(ep.Path, "/") {
			ep.Path = "/" + ep.Path
		}

		slog.Debug("adding endpoint",
			slog.String("file-path", ep.FilePath),
			slog.String("path", ep.Path))

		// assign handler to all methods
		for a := range ep.Methods {
			m := ep.Methods[a]
			mMethod, ok := methodMap[m]
			if !ok {
				slog.Warn("http method unknown",
					slog.String("endpoint", ep.FilePath),
					slog.String("method", m))

				continue
			}

			slog.Info("adding endpoint",
				slog.String("path", ep.Path),
				slog.String("method", m))

			newTree.InsertRoute(mMethod, ep.Path, ep)
		}
	}

	// replace real tree with new one
	e.lock.Lock()
	defer e.lock.Unlock()

	e.setList = endpointList
	e.currentTree = newTree
}

func MakeEndpointPluginChain(endpoint *core.Endpoint, pluginList *core.Plugins) {
	slog.Info("building plugin chain for endpoint",
		slog.String("endpoint", endpoint.FilePath))

	// warning if target not set
	if pluginList.Target.Type != "" {
		targetPlugin, err := configurePlugin(pluginList.Target,
			plugins.TargetPluginType, endpoint.Namespace)
		if err != nil {
			endpoint.Errors = append(endpoint.Errors, err.Error())
		} else {
			endpoint.TargetPluginInstance = targetPlugin
		}
	} else {
		endpoint.Warnings = append(endpoint.Warnings, "no target plugin set")
	}

	authPlugins, errors := processPlugins(pluginList.Auth,
		plugins.AuthPluginType, endpoint.Namespace)
	endpoint.AuthPluginInstances = authPlugins
	endpoint.Errors = append(endpoint.Errors, errors...)

	// inbound
	inboundPlugins, errors := processPlugins(pluginList.Inbound,
		plugins.InboundPluginType, endpoint.Namespace)
	endpoint.InboundPluginInstances = inboundPlugins
	endpoint.Errors = append(endpoint.Errors, errors...)

	// outbound
	outboundPlugins, errors := processPlugins(pluginList.Outbound,
		plugins.OutboundPluginType, endpoint.Namespace)
	endpoint.OutboundPluginInstances = outboundPlugins
	endpoint.Errors = append(endpoint.Errors, errors...)
}

func processPlugins(pluginConfigs []core.PluginConfig, t plugins.PluginType, ns string) ([]core.PluginInstance, []string) {
	errors := make([]string, 0)
	configuredPlugins := make([]core.PluginInstance, 0)

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

func configurePlugin(config core.PluginConfig, t plugins.PluginType, ns string) (core.PluginInstance, error) {
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
