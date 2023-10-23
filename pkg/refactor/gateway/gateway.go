package gateway

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/core"
)

type handler struct {
	lock *sync.Mutex

	endpoints  []*core.EndpointStatus
	pluginPool *sync.Map
}

func NewHandler() core.EndpointManager {
	return &handler{
		lock:       &sync.Mutex{},
		pluginPool: &sync.Map{},
		endpoints:  make([]*core.EndpointStatus, 0),
	}
}

func (gw *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// nolint:errcheck
	prefix := "/api/v2/gw/"
	path, _ := strings.CutPrefix(r.URL.Path, prefix)
	key := r.Method + ":/:" + path

	p, ok := gw.pluginPool.Load(key)
	if !ok {
		_, _ = w.Write([]byte("Not found!"))

		return
	}
	fList, ok := p.([]serve)
	if !ok {
		_, _ = w.Write([]byte("BUG! This should never happen."))

		return
	}
	for _, f := range fList {
		_, _ = f(w, r)
	}
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
			gw.runCycle()
			time.Sleep(time.Second)
		}

		wg.Done()
	}()
}

func (gw *handler) runCycle() {
	objectPool := &sync.Map{}
	gw.lock.Lock()
	defer gw.lock.Unlock()
	for _, endpoint := range gw.endpoints {
		res := make([]serve, 0, len(endpoint.Plugins))

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
		endpoint.Status = "healthy"
		path, _ := strings.CutPrefix(endpoint.FilePath, "/")
		objectPool.Store(endpoint.Method+":/:"+path, res)
	}
	// swap the plugin pool with the new one.
	gw.pluginPool = objectPool
}

var registry = make(map[string]Plugin)

type Plugin interface {
	build(config interface{}) (serve, error)
}

type serve func(w http.ResponseWriter, r *http.Request) (int, string)

type examplePlugin struct{}

func (examplePlugin) build(_ interface{}) (serve, error) {
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
