package endpoints

import (
	"log/slog"
	"path/filepath"
	"strings"
	"sync"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/direktiv/direktiv/pkg/refactor/spec"
)

// type EndpointEntry struct {
// 	Endpoint                *core.Endpoint
// 	AuthPluginInstances     []plugins.PluginInstance
// 	InboundPluginInstances  []plugins.PluginInstance
// 	OutboundPluginInstances []plugins.PluginInstance
// }

type EndpointList struct {
	currentTree *node

	lock sync.Mutex
}

func NewEndpointList() *EndpointList {
	return &EndpointList{
		currentTree: &node{},
	}
}

func (e *EndpointList) FindRoute(route string) (*core.Endpoint, map[string]string) {

	if !strings.HasPrefix(route, "/") {
		route = "/" + route
	}

	routeCtx := NewRouteContext()
	_, _, endpointEntry := e.currentTree.FindRoute(routeCtx, mGET, route)
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

func (e *EndpointList) SetEndpoints(endpointList []*core.Endpoint) {
	newTree := &node{}

	for i := range endpointList {
		ep := endpointList[i]

		slog.Debug("adding endpoint",
			slog.String("path", ep.FilePath),
			slog.String("extension", ep.EndpointFile.PathExtension))

		// remove the file extension, most likely .yaml
		storePath := strings.TrimSuffix(ep.FilePath, filepath.Ext(ep.FilePath))

		// add path extension if there is any
		if ep.EndpointFile.PathExtension != "" {
			storePath = filepath.Join(storePath, ep.EndpointFile.PathExtension)
		}

		auth, inbound, outbound, err := buildPluginChain(ep)
		if err != nil {
			slog.Error("configuring endpoint failed",
				slog.String("endpoint", ep.FilePath),
				slog.Any("error", err))

			continue
		}

		// create endpoint
		// entry := &EndpointEntry{
		// 	Endpoint:                ep,
		// 	AuthPluginInstances:     auth,
		// 	InboundPluginInstances:  inbound,
		// 	OutboundPluginInstances: outbound,
		// }

		// // assign handler to all methods
		// for a := range ep.EndpointFile.Methods {
		// 	m := ep.EndpointFile.Methods[a]
		// 	mMethod, ok := methodMap[m]
		// 	if !ok {
		// 		slog.Warn("http method unknown",
		// 			slog.String("endpoint", ep.FilePath),
		// 			slog.String("method", m))

		// 		continue
		// 	}

		// 	slog.Info("adding endpoint",
		// 		slog.String("path", storePath),
		// 		slog.String("method", m))

		// 	newTree.InsertRoute(mMethod, storePath, entry)
		// }
	}

	// replace real tree with new one
	e.lock.Lock()
	defer e.lock.Unlock()
	e.currentTree = newTree
}

func buildPluginChain(endpoint *core.Endpoint) error {
	// authPlugins := make([]plugins.PluginInstance, 0)
	// inboundPlugins := make([]plugins.PluginInstance, 0)
	// outboundPlugins := make([]plugins.PluginInstance, 0)

	slog.Info("building plugin chain for endpoint",
		slog.String("endpoint", endpoint.FilePath))

	createInstances := func(fc []spec.PluginConfig, pin *[]plugins.PluginInstance) error {
		pss, err := processPlugins(fc)
		if err != nil {
			slog.Error("can not process plugins", slog.String("path", endpoint.FilePath))
			return err
		}
		*pin = pss
		return nil
	}

	err := createInstances(endpoint.EndpointFile.Plugins.Auth, &endpoint.AuthPluginInstances)
	if err != nil {

	}

	// auths, err := processPlugins(endpoint.EndpointFile.Plugins.Auth)
	// if err != nil {
	// 	slog.Error("can not process auth plugins", slog.String("path", endpoint.FilePath))
	// 	return err
	// }
	// endpoint.AuthPluginInstances = auths

	// inbound, err := processPlugins(endpoint.EndpointFile.Plugins.Inboud)
	// if err != nil {
	// 	slog.Error("can not process inbound plugins", slog.String("path", endpoint.FilePath))
	// 	return err
	// }
	// endpoint.InboundPluginInstances = inbound

	// target, err := processPlugins(endpoint.EndpointFile.Plugins.Target)

	// for a := range endpoint.Plugins {
	// 	pluginDesc := endpoint.Plugins[a]

	// 	slog.Info("processing plugin",
	// 		slog.String("plugin", pluginDesc.Type))

	// 	p, err := plugins.GetPluginFromRegistry(pluginDesc.Type)
	// 	if err != nil {
	// 		return authPlugins, inboundPlugins, outboundPlugins, err
	// 	}

	// 	slog.Info("configure plugin",
	// 		slog.String("plugin", pluginDesc.Type),
	// 		slog.Any("configure", pluginDesc.Configuration))

	// 	pi, err := p.Configure(pluginDesc.Configuration)
	// 	if err != nil {
	// 		return authPlugins, inboundPlugins, outboundPlugins, err
	// 	}

	// 	switch p.Type() {
	// 	case plugins.AuthPluginType:
	// 		authPlugins = append(authPlugins, pi)
	// 	case plugins.InboundPluginType:
	// 		inboundPlugins = append(inboundPlugins, pi)
	// 	case plugins.OutboundPluginType:
	// 		outboundPlugins = append(outboundPlugins, pi)
	// 	}
	// }

	return authPlugins, inboundPlugins, outboundPlugins, nil
}

func processPlugins(pluginConfigs []spec.PluginConfig) ([]plugins.PluginInstance, error) {

	configuredPlugins := make([]plugins.PluginInstance, 0)

	for i := range pluginConfigs {

		config := pluginConfigs[i]
		slog.Info("processing plugin",
			slog.String("plugin", config.Type))

		p, err := plugins.GetPluginFromRegistry(config.Type)
		if err != nil {
			return configuredPlugins, err
		}

		pi, err := p.Configure(config.Configuration)
		if err != nil {
			return configuredPlugins, err
		}

		configuredPlugins = append(configuredPlugins, pi)
	}

	return configuredPlugins, nil
}
