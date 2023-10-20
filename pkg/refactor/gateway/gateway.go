package gateway

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/core"
)

type handler struct {
	lock *sync.Mutex

	endpoints    []*core.EndpointStatus
	pluginPool   *sync.Map
	lastChecksum string
}

func NewHandler() core.EndpointManager {
	return &handler{
		lock:      &sync.Mutex{},
		endpoints: make([]*core.EndpointStatus, 0),
	}
}

func (gw *handler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	// nolint:errcheck
	_, _ = w.Write([]byte("hello gateway"))
}

func (gw *handler) SetEndpoints(list []*core.Endpoint) {
	gw.lock.Lock()
	defer gw.lock.Unlock()

	// clone list to the gateway status.
	newList := []*core.EndpointStatus{}
	for i := range list {
		cp := *list[i]
		newList = append(newList, &core.EndpointStatus{
			Endpoint: cp,
		})
	}
	gw.endpoints = newList
}

func (gw *handler) GetAll() []*core.EndpointStatus {
	newList := []*core.EndpointStatus{}
	for i := range gw.endpoints {
		cp := *gw.endpoints[i]
		newList = append(newList, &cp)
	}

	return newList
}

func (gw *handler) Start(done <-chan struct{}, wg *sync.WaitGroup) {
	go func() {
	loop:
		for {
			select {
			case <-done:
				break loop
			default:
			}
			gw.lock.Lock()
			gw.runCycle()
			gw.lock.Unlock()
			time.Sleep(time.Second)
		}

		wg.Done()
	}()
}

func (gw *handler) runCycle() {
	objectPool := &sync.Map{}
	for _, endpoint := range gw.endpoints {
		res := make([]func(http.ResponseWriter, *http.Request) (int, string), 0, len(endpoint.Plugins))

		for _, v := range endpoint.Plugins {
			plugin, ok := registry[v.ID]
			if !ok {
				endpoint.Status = "failed"
				endpoint.Error = fmt.Sprintf("error: plugin %v not found", v.ID)
				// Empty plugin pool to prevent operating on outdated configuration.
				gw.pluginPool = &sync.Map{}

				return
			}
			servePluginFunc, err := plugin.build(v.Configuration)
			if err != nil {
				endpoint.Status = "failed"
				endpoint.Error = fmt.Sprintf("error: plugin %v configuration was rejected %v", v.ID, err)
				gw.pluginPool = &sync.Map{}

				return
			}
			// add the plugin-build.
			res = append(res, servePluginFunc)
		}
		objectPool.Store(endpoint.Namespace+":/:"+endpoint.Workflow, res)
	}
	gw.lock.Lock()
	defer gw.lock.Unlock()
	// swap the plugin pool with the new one.
	gw.pluginPool = objectPool
}

var registry = make(map[string]Plugin)

type Plugin interface {
	build(config interface{}) (func(http.ResponseWriter, *http.Request) (int, string), error)
}

type examplePlugin struct{}

func (examplePlugin) build(_ interface{}) (func(http.ResponseWriter, *http.Request) (int, string), error) {
	servePlugin := func(w http.ResponseWriter, r *http.Request) (int, string) {
		_, _ = w.Write([]byte("hello gateway"))

		return http.StatusOK, ""
	}

	return servePlugin, nil
}

//nolint:gochecknoinits
func init() {
	registry["example_plugin"] = examplePlugin{}
}
