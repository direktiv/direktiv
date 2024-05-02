package gateway2

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"sync"

	"github.com/direktiv/direktiv/pkg/refactor/gateway2/plugins"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	_ "github.com/direktiv/direktiv/pkg/refactor/gateway/plugins/auth"
	_ "github.com/direktiv/direktiv/pkg/refactor/gateway/plugins/inbound"
	_ "github.com/direktiv/direktiv/pkg/refactor/gateway/plugins/outbound"
	_ "github.com/direktiv/direktiv/pkg/refactor/gateway/plugins/target"
)

const ctxKeyConsumers = "ctx_consumers"

type manager struct {
	mux       *sync.Mutex
	router    *http.ServeMux
	endpoints []core.EndpointV2
	consumers []core.ConsumerV2
}

func NewManager() core.GatewayManagerV2 {
	return &manager{
		mux:       &sync.Mutex{},
		endpoints: make([]core.EndpointV2, 0),
		consumers: make([]core.ConsumerV2, 0),
	}
}

func (m *manager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mux.Lock()
	defer m.mux.Unlock()

	m.router.ServeHTTP(w, r)
}

func (m *manager) SetEndpoints(list []core.EndpointV2) {
	m.mux.Lock()
	defer m.mux.Unlock()

	m.endpoints = list
	m.build()
}

func (m *manager) SetConsumers(list []core.ConsumerV2) {
	m.mux.Lock()
	defer m.mux.Unlock()

	m.consumers = list
	m.build()
}

func (m *manager) build() {
	newRouter := http.NewServeMux()

	// reset all errors.
	for i := range m.endpoints {
		m.endpoints[i].Errors = []error{}
	}

	for i, item := range m.endpoints {
		// concat plugins configs into one list.
		pConfigs := []core.PluginConfigV2{}
		pConfigs = append(pConfigs, item.PluginsConfig.Auth...)
		pConfigs = append(pConfigs, item.PluginsConfig.Inbound...)
		pConfigs = append(pConfigs, item.PluginsConfig.Target)
		pConfigs = append(pConfigs, item.PluginsConfig.Outbound...)

		// build plugins chain.
		pChain := []core.PluginV2{}
		for _, pConfig := range pConfigs {
			p, err := plugins.NewPlugin(pConfig)
			if err != nil {
				item.Errors = append(item.Errors, err)
			}
			pChain = append(pChain, p)
		}

		// only mount http handler when plugins has zero errors.
		if len(item.Errors) == 0 {
			newRouter.HandleFunc(item.Path, func(w http.ResponseWriter, r *http.Request) {
				// check if correct method.
				if !slices.Contains(item.Methods, r.Method) {
					writeJsonError(w, http.StatusMethodNotAllowed, item.FilePath,
						fmt.Sprintf("method:%s is not allowed with this endpoint", r.Method))
				}

				// inject consumer files.
				r = r.WithContext(context.WithValue(r.Context(), ctxKeyConsumers,
					m.listNamespacedConsumers(item.Namespace)))

				for _, p := range pChain {
					if ok := p.Execute(w, r); !ok {
						break
					}
				}
			})
		}

		m.endpoints[i] = item
	}

	// set the new router.
	m.router = newRouter
}

func (m *manager) listNamespacedConsumers(namespace string) []core.ConsumerV2 {
	list := []core.ConsumerV2{}
	for _, item := range m.consumers {
		if item.Namespace == namespace {
			list = append(list, item)
		}
	}

	return list
}

func (m *manager) listNamespacedEndpoints(namespace string) []core.EndpointV2 {
	list := []core.EndpointV2{}
	for _, item := range m.endpoints {
		if item.Namespace == namespace {
			list = append(list, item)
		}
	}

	return list
}

func (m *manager) ListEndpoints(namespace string) []core.EndpointV2 {
	m.mux.Lock()
	defer m.mux.Unlock()

	return m.listNamespacedEndpoints(namespace)
}

func (m *manager) ListConsumers(namespace string) []core.ConsumerV2 {
	m.mux.Lock()
	defer m.mux.Unlock()

	return m.listNamespacedConsumers(namespace)
}

func writeJsonError(w http.ResponseWriter, status int, endpointFile string, err string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	payload := struct {
		EndpointFile string `json:"endpointFile"`
		Error        any    `json:"error"`
	}{
		EndpointFile: endpointFile,
		Error:        err,
	}
	_ = json.NewEncoder(w).Encode(payload)
}
