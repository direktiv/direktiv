package gateway2

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/mitchellh/mapstructure"
)

const (
	ConsumerUserHeader   = "Direktiv-Consumer-User"
	ConsumerTagsHeader   = "Direktiv-Consumer-Tags"
	ConsumerGroupsHeader = "Direktiv-Consumer-Groups"
)

func ConvertConfig(config any, target any) error {
	err := mapstructure.Decode(config, target)
	if err != nil {
		return errors.Join(err, errors.New("configuration invalid"))
	}

	return nil
}

func hasActiveConsumer(r *http.Request) bool {
	return r.Context().Value(core.GatewayCtxKeyActiveConsumer) != nil
}

func isAuthPlugin(p core.PluginV2) bool {
	return strings.Contains(p.Type(), "-auth") || strings.Contains(p.Type(), "auth-")
}

func WriteJSONError(w http.ResponseWriter, status int, endpointFile string, msg string) {
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

func WriteJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	payLoad := struct {
		Data any `json:"data"`
	}{
		Data: v,
	}
	_ = json.NewEncoder(w).Encode(payLoad)
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
