package gateway2

import (
	"context"
	"fmt"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2/plugins"
	"net/http"
	"slices"
	"strings"
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
