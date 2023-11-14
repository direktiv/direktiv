package gateway

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/go-chi/chi/v5"
	"github.com/invopop/jsonschema"
)

var registry = make(map[string]plugin)

type plugin interface {
	build(config map[string]interface{}) (serve, error)
	getSchema() interface{}
}

type serve func(w http.ResponseWriter, r *http.Request) bool

type handler struct {
	pluginPool map[string]endpointEntry
	mu         sync.Mutex
}

type endpointEntry struct {
	*core.EndpointStatus
	item     int
	checksum string
	plugins  []serve
}

func NewHandler() core.EndpointManager {
	return &handler{}
}

func (gw *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	routePath := chi.URLParam(r, "*")

	fmt.Printf("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! %v\n", routePath)
	prefix := "/api/v2/gw/"
	path, _ := strings.CutPrefix(r.URL.Path, prefix)
	key := r.Method + ":/:" + path

	endpoint, ok := gw.pluginPool[key]
	if !ok {
		http.NotFound(w, r)

		return
	}
	for _, f := range endpoint.plugins {
		cont := f(w, r)
		if !cont {
			return
		}
	}
}

func (gw *handler) SetEndpoints(list []*core.Endpoint) {
	gw.mu.Lock() // Lock
	defer gw.mu.Unlock()
	// newList := make([]*core.EndpointStatus, len(list))
	// newPool := make(map[string]endpointEntry)
	// oldPool := gw.pluginPool

	for i := range list {

		// flow.sugar.Error(fmt.Sprintf("deleting all logs since %v", t))
		fmt.Printf("COUHTER %v\n", i)

		slog.Info("hello", "count", i)
		// cp := *ep
		// newList[i] = &core.EndpointStatus{Endpoint: cp}

		// path, _ := strings.CutPrefix(cp.FilePath, "/")
		// checksum := string(sha256.New().Sum([]byte(fmt.Sprint(cp))))

		// key := cp.Method + ":/:" + path

		// fmt.Printf("!!!!!!!!!!!!!!!!!!! %v\n", key)
		// value, ok := oldPool[key]
		// if ok && value.checksum == checksum {
		// 	newPool[key] = value
		// }
		// newPool[key] = endpointEntry{
		// 	plugins:        buildPluginChain(newList[i]),
		// 	EndpointStatus: newList[i],
		// 	checksum:       checksum,
		// 	item:           i,
		// }
	}
	// gw.pluginPool = newPool
}

func buildPluginChain(endpoint *core.EndpointStatus) []serve {
	res := make([]serve, 0, len(endpoint.Plugins))

	for _, v := range endpoint.Plugins {
		plugin, ok := registry[v.Type]
		if !ok {
			endpoint.Error = fmt.Sprintf("error: plugin %v not found", v.Type)

			continue
		}

		servePluginFunc, err := plugin.build(v.Configuration)
		if err != nil {
			endpoint.Error = fmt.Sprintf("error: plugin %v configuration was rejected %v", v.Type, err)

			continue
		}

		res = append(res, servePluginFunc)
	}

	return res
}

func (gw *handler) GetAll() []*core.EndpointStatus {
	gw.mu.Lock() // Lock
	defer gw.mu.Unlock()

	newList := make([]*core.EndpointStatus, len(gw.pluginPool))

	for _, value := range gw.pluginPool {
		newList[value.item] = value.EndpointStatus
	}

	return newList
}

func GetAllSchemas() (any, error) {
	res := make(map[string]interface{})

	for k, p := range registry {
		schemaStruct := p.getSchema()
		schema := jsonschema.Reflect(schemaStruct)

		var schemaObj map[string]interface{}
		b, err := schema.MarshalJSON()
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(b, &schemaObj); err != nil {
			return nil, err
		}

		res[k] = schemaObj
	}

	return res, nil
}
