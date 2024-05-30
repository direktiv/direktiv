package gateway2

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/core"
)

// Struct router implements the gateway logic of serving requests. We can see that it wraps a simple
// http.ServeMux with endpoints and consumers. Lists  endpoints and consumers are used to build the router itself.
type router struct {
	serveMux  *http.ServeMux
	endpoints []core.EndpointV2
	consumers []core.ConsumerV2
}

// buildRouter compiles a new gateway router from endpoints and consumers lists.
//
//nolint:gocognit
func buildRouter(endpoints []core.EndpointV2, consumers []core.ConsumerV2) *router {
	serveMux := http.NewServeMux()

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
			p, err := NewPlugin(pConfig)
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
		serveMux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			// check if correct method.
			if !slices.Contains(item.Methods, r.Method) {
				WriteJSONError(w, http.StatusMethodNotAllowed, item.FilePath,
					fmt.Sprintf("method:%s is not allowed with this endpoint", r.Method))

				return
			}

			// inject consumer files.
			r = InjectContextConsumersList(r, filterNamespacedConsumers(consumers, item.Namespace))
			// inject endpoint.
			r = InjectContextEndpoint(r, &endpoints[i])

			for _, p := range pChain {
				// checkpoint if auth plugins had a match.
				if !isAuthPlugin(p) {
					// case where auth is required but request is not authenticated (consumers doesn't match).
					hasActiveConsumer := ExtractContextActiveConsumer(r) != nil
					if !item.AllowAnonymous && !hasActiveConsumer {
						WriteJSONError(w, http.StatusForbidden, item.FilePath, "authentication failed")

						return
					}
				}
				if r = p.Execute(w, r); r == nil {
					break
				}
			}
		})
	}

	// mount not found route.
	serveMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		WriteJSONError(w, http.StatusNotFound, "", "gateway couldn't find a matching endpoint")
	})

	return &router{
		serveMux:  serveMux,
		endpoints: make([]core.EndpointV2, 0),
		consumers: make([]core.ConsumerV2, 0),
	}
}
