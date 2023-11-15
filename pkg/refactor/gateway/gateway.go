package gateway

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/go-chi/chi/v5"

	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	_ "github.com/direktiv/direktiv/pkg/refactor/gateway/plugins/auth"
)

var pathTree = &node{}

type handler struct {
	pluginPool map[string]endpointEntry

	routeLock sync.Mutex
}

type endpointEntry struct {
	endpoint        *core.Endpoint
	authPlugins     []plugins.Plugin
	inboundPlugins  []plugins.Plugin
	outboundPlugins []plugins.Plugin
}

func NewHandler() core.EndpointManager {
	return &handler{}
}

func (gw *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	routePath := chi.URLParam(r, "*")

	ctx := NewRouteContext()
	_, _, endpointEntry := pathTree.FindRoute(ctx, mGET, "/"+routePath)

	if endpointEntry == nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
		return
	}

	// TODO: set timeout
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Printf("!!!!!!!!!!!!!!!!!!!!>>>1 %v\n", endpointEntry)

	fmt.Printf("!!!!!!!!!!!!!!!!!!!!>>>2 %v\n", endpointEntry.authPlugins)

	// run auth
	for i := range endpointEntry.authPlugins {
		authPlugin := endpointEntry.authPlugins[i]
		authPlugin.ExecutePlugin(ctxTimeout, w, r)

		// chech if consumer is set in plugin context
	}

	fmt.Printf("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! %v\n", endpointEntry.endpoint.FilePath)
	// prefix := "/api/v2/gw/"
	// path, _ := strings.CutPrefix(r.URL.Path, prefix)
	// key := r.Method + ":/:" + path

	// endpoint, ok := gw.pluginPool[key]
	// if !ok {
	// 	http.NotFound(w, r)

	// 	return
	// }
	// for _, f := range endpoint.plugins {
	// 	cont := f(w, r)
	// 	if !cont {
	// 		return
	// 	}
	// }
}

func (gw *handler) SetEndpoints(endpointList []*core.Endpoint) {

	var newTree = &node{}

	for i := range endpointList {

		ep := endpointList[i]

		slog.Debug("adding endpoint",
			slog.String("path", ep.FilePath),
			slog.String("extension", ep.PathExtension))

		// remove the file extension, most likely .yaml
		storePath := strings.TrimSuffix(ep.FilePath, filepath.Ext(ep.FilePath))

		// add path extension if there is any
		if ep.PathExtension != "" {
			storePath = filepath.Join(storePath, ep.PathExtension)
		}

		auth, inbound, outbound, err := buildPluginChain(ep)
		if err != nil {
			slog.Error("configuring endpoint failed",
				slog.String("endpoint", ep.FilePath),
				slog.Any("error", err))
			continue
		}

		// create endpoint
		entry := &endpointEntry{
			endpoint:        ep,
			authPlugins:     auth,
			inboundPlugins:  inbound,
			outboundPlugins: outbound,
		}

		fmt.Printf("!!!!!!!!!!!!!!!!!!!!>>> %v\n", auth)

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
				slog.String("path", storePath),
				slog.String("method", m))

			newTree.InsertRoute(mMethod, storePath, entry)
		}
	}

	// replace real tree with new one
	gw.routeLock.Lock()
	defer gw.routeLock.Unlock()
	pathTree = newTree

	// plugin, err := plugins.GetPluginFromRegistry("basic-auth")

	// fmt.Println(err)
	// fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> ")

	// fmt.Println(plugin.Name())
	// fmt.Println(plugin)

	// config := make(map[string]interface{})

	// config["dsd"] = "sdd"

	// fmt.Println(plugin.Configure(config))

	// config["val2"] = "val3"
	// fmt.Println(plugin.Configure(config))

	// slog.Info("sss", "chidlren", pathTree.children)

	// ctx := &Context{}
	// // ctx.RoutePath = "/"
	// a, b, c := pathTree.FindRoute(ctx, mGET, "/endpoints/image/123")

	// fmt.Printf("%+v %v %v\n", a, b, c)

	// fmt.Printf("%+v\n", ctx)

	// for k := range a.endpoints {
	// 	ep := a.endpoints[k]
	// 	fmt.Println(ep.handler)
	// 	fmt.Println(ep.paramKeys)
	// 	fmt.Println(ep.pattern)
	// }

	// slog.Info("sss", "chidlren", newTree.routes())
}

func buildPluginChain(endpoint *core.Endpoint) ([]plugins.Plugin, []plugins.Plugin, []plugins.Plugin, error) {

	authPlugins := make([]plugins.Plugin, 0)
	inboundPlugins := make([]plugins.Plugin, 0)
	outboundPlugins := make([]plugins.Plugin, 0)

	slog.Info("building plugin chain for endpoint",
		slog.String("endpoint", endpoint.FilePath))

	for a := range endpoint.Plugins {
		pluginDesc := endpoint.Plugins[a]

		slog.Info("processing plugin",
			slog.String("plugin", pluginDesc.Type))

		p, err := plugins.GetPluginFromRegistry(pluginDesc.Type)
		if err != nil {
			return authPlugins, inboundPlugins, outboundPlugins, err
		}

		pi, err := p.Configure(pluginDesc.Configuration)
		if err != nil {
			return authPlugins, inboundPlugins, outboundPlugins, err
		}

		switch p.Type() {
		case plugins.AuthPluginType:
			authPlugins = append(authPlugins, pi)
		case plugins.InboundPluginType:
			inboundPlugins = append(inboundPlugins, pi)
		case plugins.OutboundPluginType:
			outboundPlugins = append(outboundPlugins, pi)
		}
	}

	slog.Info("processing plugin!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!",
		slog.Any("plugin", authPlugins))

	return authPlugins, inboundPlugins, outboundPlugins, nil
}

// 	res := make([]plugins.Serve, 0, len(endpoint.Plugins))

// 	// for _, v := range endpoint.Plugins {
// 	// 	plugin, ok := registry[v.Type]
// 	// 	if !ok {
// 	// 		endpoint.Error = fmt.Sprintf("error: plugin %v not found", v.Type)

// 	// 		continue
// 	// 	}

// 	// 	servePluginFunc, err := plugin.build(v.Configuration)
// 	// 	if err != nil {
// 	// 		endpoint.Error = fmt.Sprintf("error: plugin %v configuration was rejected %v", v.Type, err)

// 	// 		continue
// 	// 	}

// 	// 	res = append(res, servePluginFunc)
// 	// }

// 	return res
// }

func (gw *handler) GetAll() []*core.Endpoint {
	// gw.mu.Lock() // Lock
	// defer gw.mu.Unlock()

	newList := make([]*core.Endpoint, len(gw.pluginPool))

	// for _, value := range gw.pluginPool {
	// 	// newList[value.item] = value.Endpoint
	// }

	return newList
}

func GetAllSchemas() (any, error) {
	res := make(map[string]interface{})

	// for k, p := range registry {
	// 	schemaStruct := p.getSchema()
	// 	schema := jsonschema.Reflect(schemaStruct)

	// 	var schemaObj map[string]interface{}
	// 	b, err := schema.MarshalJSON()
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	if err := json.Unmarshal(b, &schemaObj); err != nil {
	// 		return nil, err
	// 	}

	// 	res[k] = schemaObj
	// }

	return res, nil
}
