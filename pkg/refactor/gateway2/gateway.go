package gateway2

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2/plugins"
)

type immutableManager struct {
	router    *http.ServeMux
	endpoints []core.EndpointV2
	consumers []core.ConsumerV2
}

func newManager(endpoints []core.EndpointV2, consumers []core.ConsumerV2) *immutableManager {
	newRouter := http.NewServeMux()

	for i, item := range endpoints {
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
				item.Errors = append(item.Errors, fmt.Errorf("plugin '%s' config: %w", pConfig.Typ, err))
			}
			pChain = append(pChain, p)
		}
		if len(item.PluginsConfig.Auth) == 0 && !item.AllowAnonymous {
			item.Errors = append(item.Errors, fmt.Errorf("AllowAnonymous is false but zero auth plugin configured"))
		}
		endpoints[i] = item

		// skip mount http handler when plugins has zero errors.
		if len(item.Errors) > 0 {
			continue
		}

		cleanPath := strings.Trim(item.Path, " /")
		pattern := fmt.Sprintf("/api/v2/namespaces/%s/gateway2/%s", item.Namespace, cleanPath)
		newRouter.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			// check if correct method.
			if !slices.Contains(item.Methods, r.Method) {
				writeJSONError(w, http.StatusMethodNotAllowed, item.FilePath,
					fmt.Sprintf("method:%s is not allowed with this endpoint", r.Method))

				return
			}
			// inject consumer files.
			r = r.WithContext(context.WithValue(r.Context(), core.GatewayCtxKeyConsumers,
				filterNamespacedConsumers(consumers, item.Namespace)))

			var err error
			for _, p := range pChain {
				// checkpoint if auth plugins had a match.
				if !isAuthPlugin(p) {
					// case where auth is required but request is not authenticated (consumers doesn't match).
					if !item.AllowAnonymous && !hasActiveConsumer(r) {
						writeJSONError(w, http.StatusForbidden, item.FilePath, "authentication failed")

						return
					}
				}
				if r, err = p.Execute(w, r); err != nil {
					writeJSONError(w, http.StatusInternalServerError, item.FilePath, fmt.Sprintf("gateway plugin(%s) execution failed", p.Type()))
					// TODO: verbose log here.
					break
				}
			}
		})
	}

	// mount not found route.
	newRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeJSONError(w, http.StatusNotFound, "", "gateway couldn't find a matching endpoint")
	})

	return &immutableManager{
		router:    newRouter,
		endpoints: make([]core.EndpointV2, 0),
		consumers: make([]core.ConsumerV2, 0),
	}
}

func filterNamespacedConsumers(consumers []core.ConsumerV2, namespace string) []core.ConsumerV2 {
	list := []core.ConsumerV2{}
	for _, item := range consumers {
		if item.Namespace == namespace {
			list = append(list, item)
		}
	}

	return list
}

func filterNamespacedEndpoints(endpoints []core.EndpointV2, namespace string) []core.EndpointV2 {
	list := []core.EndpointV2{}
	for _, item := range endpoints {
		if item.Namespace == namespace {
			list = append(list, item)
		}
	}

	return list
}

type manager struct {
	inner *immutableManager
}

var _ core.GatewayManagerV2 = &manager{}

func (m *manager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if m.inner == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "", "no active gateway endpoints")

		return
	}
	m.inner.router.ServeHTTP(w, r)
}

func (m *manager) SetEndpoints(list []core.EndpointV2, cList []core.ConsumerV2) {
	newOne := newManager(list, cList)
	m.inner = newOne
}

func (m *manager) ListEndpoints(namespace string) []core.EndpointV2 {
	return filterNamespacedEndpoints(m.inner.endpoints, namespace)
}

func (m *manager) ListConsumers(namespace string) []core.ConsumerV2 {
	return filterNamespacedConsumers(m.inner.consumers, namespace)
}

func writeJSONError(w http.ResponseWriter, status int, endpointFile string, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	inner := struct {
		EndpointFile string `json:"endpointFile,omitempty"`
		Message      any    `json:"message"`
	}{
		EndpointFile: endpointFile,
		Message:      msg,
	}
	payload := struct {
		Error any `json:"error"`
	}{
		Error: inner,
	}

	_ = json.NewEncoder(w).Encode(payload)
}

func isAuthPlugin(p core.PluginV2) bool {
	return strings.Contains(p.Type(), "-auth") || strings.Contains(p.Type(), "auth-")
}

func hasActiveConsumer(r *http.Request) bool {
	return r.Context().Value(core.GatewayCtxKeyActiveConsumer) != nil
}
