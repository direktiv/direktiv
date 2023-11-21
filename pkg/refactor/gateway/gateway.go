package gateway

import (
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	// This triggers the init function within for auth plugins to register them.

	"github.com/direktiv/direktiv/pkg/refactor/gateway/consumer"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/endpoints"
	_ "github.com/direktiv/direktiv/pkg/refactor/gateway/plugins/auth"

	// This triggers the init function within for inbound plugins to register them.
	_ "github.com/direktiv/direktiv/pkg/refactor/gateway/plugins/inbound"
)

type handler struct {
	// pluginPool map[string]endpointEntry

	EndpointList *endpoints.EndpointList
	ConsumerList *consumer.ConsumerList

	// routeLock sync.Mutex
}

func NewHandler() core.EndpointManager {
	return &handler{
		// EndpointList: endpoints.NewEndpointList(),
		// ConsumerList: consumer.NewConsumerList(),
	}
}

func (gw *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// routePath := chi.URLParam(r, "*")

	// routeCtx := NewRouteContext()
	// _, _, endpointEntry := pathTree.FindRoute(routeCtx, mGET, "/"+routePath)

	// // add path extension variables in context, e.g. /{id}
	// urlParams := make(map[string]string)
	// for i := 0; i < len(routeCtx.URLParams.Keys); i++ {
	// 	key := routeCtx.URLParams.Keys[i]
	// 	urlParams[key] = routeCtx.URLParams.Values[i]
	// }
	// ctx := context.WithValue(r.Context(), plugins.URLParamCtxKey, urlParams)

	// if endpointEntry == nil {
	// 	w.WriteHeader(http.StatusNotFound)
	// 	// nolint
	// 	w.Write([]byte("not found"))

	// 	return
	// }

	// // TODO: set timeout
	// ctxTimeout, cancel := context.WithTimeout(ctx, time.Second)
	// defer cancel()

	// c := &core.Consumer{}

	// // run auth
	// for i := range endpointEntry.authPlugins {
	// 	authPlugin := endpointEntry.authPlugins[i]
	// 	authPlugin.ExecutePlugin(ctxTimeout, c, w, r)

	// 	// check and exit if consumer is set in plugin
	// 	if c.Username != "" {
	// 		slog.Info("user authenticated", "user", c.Username)
	// 		break
	// 	}
	// }

	// // if user not authenticated and anonymous access not enabled
	// if c.Username == "" && !endpointEntry.endpoint.AllowAnonymous {
	// 	w.WriteHeader(http.StatusUnauthorized)
	// 	// nolint
	// 	w.Write([]byte("unauthorized"))
	// }

	// for i := range endpointEntry.inboundPlugins {
	// 	inboundPlugin := endpointEntry.inboundPlugins[i]
	// 	result := inboundPlugin.ExecutePlugin(ctxTimeout, c, w, r)
	// 	if !result {
	// 		fmt.Println("DONT!")
	// 	}
	// }

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

	// newList := make([]*core.Endpoint, len(gw.pluginPool))

	// for _, value := range gw.pluginPool {
	// 	// newList[value.item] = value.Endpoint
	// }

	// return newList

	return nil
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
