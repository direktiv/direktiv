package gateway

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/direktiv/direktiv/pkg/core"
)

const (
	ConsumerUserHeader   = "Direktiv-Consumer-User"
	ConsumerTagsHeader   = "Direktiv-Consumer-Tags"
	ConsumerGroupsHeader = "Direktiv-Consumer-Groups"
)

func isAuthPlugin(p core.Plugin) bool {
	return strings.Contains(p.Type(), "-auth") || strings.Contains(p.Type(), "auth-")
}

func filterNamespacedConsumers(consumers []core.Consumer, namespace string) []core.Consumer {
	list := []core.Consumer{}
	for _, item := range consumers {
		if item.Namespace == namespace {
			list = append(list, item)
		}
	}

	return list
}

func filterNamespacedEndpoints(endpoints []core.Endpoint, namespace string, path string) []core.Endpoint {
	list := []core.Endpoint{}
	for _, item := range endpoints {
		if item.Namespace == namespace && (path == "" || path == item.Path) {
			list = append(list, item)
		}
	}

	return list
}

// FindConsumerByUser find a consumer that matches a user string.
func FindConsumerByUser(list []core.Consumer, user string) *core.Consumer {
	for _, item := range list {
		if item.Username == user {
			return &item
		}
	}

	return nil
}

// FindConsumerByAPIKey find a consumer that matches a key string.
func FindConsumerByAPIKey(list []core.Consumer, key string) *core.Consumer {
	for _, item := range list {
		if item.APIKey == key {
			return &item
		}
	}

	return nil
}

// WriteJSON helper function to write a json payload to a http.ResponseWriter.
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
