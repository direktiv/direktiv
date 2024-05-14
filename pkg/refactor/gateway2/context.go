package gateway2

import (
	"context"
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/core"
)

const (
	gatewayCtxKeyConsumersList  = "ctx_consumers_list"
	gatewayCtxKeyActiveConsumer = "ctx_active_consumer"
	gatewayCtxKeyNamespace      = "ctx_namespace"
	gatewayCtxKeyEndpoint       = "ctx_endpoint"
)

func ExtractContextConsumersList(r *http.Request) []core.ConsumerV2 {
	res := r.Context().Value(gatewayCtxKeyConsumersList)
	if res == nil {
		return nil
	}
	cast, ok := res.([]core.ConsumerV2)
	if !ok {
		return nil
	}

	return cast
}

func InjectContextConsumersList(r *http.Request, contextValue []core.ConsumerV2) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), gatewayCtxKeyConsumersList, contextValue))
}

func ExtractContextActiveConsumer(r *http.Request) *core.ConsumerV2 {
	res := r.Context().Value(gatewayCtxKeyActiveConsumer)
	if res == nil {
		return nil
	}
	cast, ok := res.(*core.ConsumerV2)
	if !ok {
		return nil
	}

	return cast
}

func InjectContextActiveConsumer(r *http.Request, contextValue *core.ConsumerV2) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), gatewayCtxKeyActiveConsumer, contextValue))
}

func ExtractContextNamespace(r *http.Request) string {
	res := r.Context().Value(gatewayCtxKeyNamespace)
	if res == nil {
		return ""
	}
	cast, ok := res.(string)
	if !ok {
		return ""
	}

	return cast
}

func InjectContextNamespace(r *http.Request, contextValue string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), gatewayCtxKeyNamespace, contextValue))
}

func ExtractContextEndpoint(r *http.Request) *core.EndpointV2 {
	res := r.Context().Value(gatewayCtxKeyEndpoint)
	if res == nil {
		return nil
	}
	cast, ok := res.(*core.EndpointV2)
	if !ok {
		return nil
	}

	return cast
}

func InjectContextEndpoint(r *http.Request, contextValue *core.EndpointV2) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), gatewayCtxKeyEndpoint, contextValue))
}
