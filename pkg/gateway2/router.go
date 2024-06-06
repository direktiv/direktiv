package gateway2

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"slices"
	"strings"

	"github.com/direktiv/direktiv/pkg/core"
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
		// don't process endpoints with errors
		if len(item.Errors) > 0 {
			continue
		}

		// concat plugins configs into one list.
		pConfigs := []core.PluginConfigV2{}
		pConfigs = append(pConfigs, item.PluginsConfig.Auth...)
		pConfigs = append(pConfigs, item.PluginsConfig.Inbound...)
		pConfigs = append(pConfigs, item.PluginsConfig.Target)
		pConfigs = append(pConfigs, item.PluginsConfig.Outbound...)

		hasOutboundConfigured := len(item.PluginsConfig.Outbound) > 0

		// build plugins chain.
		pChain := []core.PluginV2{}
		for _, pConfig := range pConfigs {
			p, err := NewPlugin(pConfig)
			if err != nil {
				item.Errors = append(item.Errors, fmt.Sprintf("plugin '%s' err: %s", pConfig.Typ, err))
			}
			pChain = append(pChain, p)
		}
		if len(item.PluginsConfig.Auth) == 0 && !item.AllowAnonymous {
			item.Errors = append(item.Errors, "AllowAnonymous is false but zero auth plugin configured")
		}
		endpoints[i] = item

		// skip mount http handler when plugins has zero errors.
		if len(item.Errors) > 0 {
			continue
		}

		cleanPath := strings.Trim(item.Path, " /")

		for _, pattern := range []string{
			fmt.Sprintf("/api/v2/namespaces/%s/gateway/%s", item.Namespace, cleanPath),
			fmt.Sprintf("/ns/%s/%s", item.Namespace, cleanPath),
		} {
			serveMux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
				// Check if correct method.
				if !slices.Contains(item.Methods, r.Method) {
					WriteJSONError(w, http.StatusMethodNotAllowed, item.FilePath,
						fmt.Sprintf("method:%s is not allowed with this endpoint", r.Method))

					return
				}

				// Inject consumer files.
				r = InjectContextConsumersList(r, filterNamespacedConsumers(consumers, item.Namespace))
				// inject endpoint.
				r = InjectContextEndpoint(r, &endpoints[i])
				r = InjectContextURLParams(r, ExtractBetweenCurlyBraces(pattern))

				// Outbound plugins are used to transform the output from target plugins. When an outbound plugin is
				// configured, target plugins output should be recorded in a buffer rather than flushed directly to
				// the client's tcp connection. Then the recorded bytes should be somehow piped to the outbound
				// plugins.
				originalWriter := w
				if hasOutboundConfigured {
					w = httptest.NewRecorder()
				}

				for _, p := range pChain {
					// Checkpoint if auth plugins had a match.
					if !isAuthPlugin(p) {
						// Case where auth is required but request is not authenticated (consumers doesn't match).
						hasActiveConsumer := ExtractContextActiveConsumer(r) != nil
						if !item.AllowAnonymous && !hasActiveConsumer {
							WriteJSONError(w, http.StatusForbidden, item.FilePath, "authentication failed")

							break
						}
					}
					if p.Type() == "js-outbound" {
						// Inject the output in the request so that the outbound plugin can process it.
						//nolint:forcetypeassert
						w := w.(*httptest.ResponseRecorder)
						newReq, err := http.NewRequest(http.MethodGet, "/writer", w.Body)
						if err != nil {
							slog.With("component", "gateway").
								Error("creating js-outbound plugin request", "err", err)
						}
						newReq.Response = &http.Response{
							StatusCode: w.Code,
						}
						//nolint:contextcheck
						newReq = newReq.WithContext(r.Context())
						r = newReq
					}
					if r = p.Execute(w, r); r == nil {
						break
					}
				}

				if hasOutboundConfigured {
					//nolint:forcetypeassert
					w := w.(*httptest.ResponseRecorder)
					// Copy headers to the original writer.
					for key, values := range w.Header() {
						for _, value := range values {
							originalWriter.Header().Add(key, value)
						}
					}
					// Copy status code to the original writer.
					originalWriter.WriteHeader(w.Code)

					// Copy body to the original writer.
					if _, err := io.Copy(originalWriter, w.Body); err != nil {
						slog.With("component", "gateway").
							Error("flushing final bytes to connection", "err", err)
					}
				}
			})
		}
	}

	// Mount not found route
	serveMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		WriteJSONError(w, http.StatusNotFound, "", "gateway couldn't find a matching endpoint")
	})

	return &router{
		serveMux:  serveMux,
		endpoints: endpoints,
		consumers: consumers,
	}
}

func ExtractBetweenCurlyBraces(input string) []string {
	// Compile the regular expression
	re := regexp.MustCompile(`\{([^{}]*)\}`)

	// Find all matches
	matches := re.FindAllStringSubmatch(input, -1)

	// Extract the matched strings
	var results []string
	for _, match := range matches {
		// match[0] is the full match (e.g., "{example}")
		// match[1] is the first capturing group (e.g., "example")
		if len(match) > 1 {
			results = append(results, match[1])
		}
	}

	return results
}
