package gateway

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/direktiv/direktiv/pkg/refactor/core"
)

var registry = make(map[string]Plugin)

type Plugin interface {
	build(config interface{}) (serve, error)
}

type serve func(w http.ResponseWriter, r *http.Request) bool

type examplePlugin struct{}

func (examplePlugin) build(_ interface{}) (serve, error) {
	return func(w http.ResponseWriter, r *http.Request) bool {
		_, _ = w.Write([]byte("hello gateway"))

		return true
	}, nil
}

//nolint:gochecknoinits
func init() {
	registry["example_plugin"] = examplePlugin{}
}

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
	prefix := "/api/v2/gw/"
	path, _ := strings.CutPrefix(r.URL.Path, prefix)
	key := r.Method + ":/:" + path

	fList, ok := gw.pluginPool[key]
	if !ok {
		_, _ = w.Write([]byte("Not found!"))

		return
	}

	for _, f := range fList.plugins {
		cont := f(w, r)
		if !cont {
			return
		}
	}
}

func (gw *handler) SetEndpoints(list []*core.Endpoint) []*core.EndpointStatus {
	gw.mu.Lock() // Lock
	defer gw.mu.Unlock()
	newList := make([]*core.EndpointStatus, len(list))
	newPool := make(map[string]endpointEntry)
	oldPool := gw.pluginPool

	for i, ep := range list {
		cp := *ep
		newList[i] = &core.EndpointStatus{Endpoint: cp}

		path, _ := strings.CutPrefix(cp.FilePath, "/")
		checksum := string(sha256.New().Sum([]byte(fmt.Sprint(cp))))

		key := cp.Method + ":/:" + path
		value, ok := oldPool[key]
		if ok && value.checksum == checksum {
			newPool[key] = value
		}
		newList[i].Status = "applied"
		newPool[key] = endpointEntry{
			plugins:        buildPluginChain(newList[i]),
			EndpointStatus: newList[i],
			checksum:       checksum,
			item:           i,
		}
	}
	gw.pluginPool = newPool

	return newList
}

func buildPluginChain(endpoint *core.EndpointStatus) []serve {
	res := make([]serve, 0, len(endpoint.Plugins))

	for _, v := range endpoint.Plugins {
		plugin, ok := registry[v.ID]
		if !ok {
			endpoint.Status = "failed"
			endpoint.Error = fmt.Sprintf("error: plugin %v not found", v.ID)

			continue
		}

		servePluginFunc, err := plugin.build(v.Configuration)
		if err != nil {
			endpoint.Status = "failed"
			endpoint.Error = fmt.Sprintf("error: plugin %v configuration was rejected %v", v.ID, err)

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
