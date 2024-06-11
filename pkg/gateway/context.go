package gateway

import (
	"context"
	"net/http"

	"github.com/direktiv/direktiv/pkg/core"
)

const (
	gatewayCtxKeyConsumersList  = "ctx_consumers_list"
	gatewayCtxKeyActiveConsumer = "ctx_active_consumer"
	gatewayCtxKeyEndpoint       = "ctx_endpoint"
	gatewayCtxKeyURLParams      = "ctx_params"
)

func ExtractContextConsumersList(r *http.Request) []core.Consumer {
	res := r.Context().Value(gatewayCtxKeyConsumersList)
	if res == nil {
		return nil
	}
	cast, ok := res.([]core.Consumer)
	if !ok {
		return nil
	}

	return cast
}

func InjectContextConsumersList(r *http.Request, contextValue []core.Consumer) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), gatewayCtxKeyConsumersList, contextValue))
}

func ExtractContextActiveConsumer(r *http.Request) *core.Consumer {
	res := r.Context().Value(gatewayCtxKeyActiveConsumer)
	if res == nil {
		return nil
	}
	cast, ok := res.(*core.Consumer)
	if !ok {
		return nil
	}

	return cast
}

func InjectContextActiveConsumer(r *http.Request, contextValue *core.Consumer) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), gatewayCtxKeyActiveConsumer, contextValue))
}

func ExtractContextEndpoint(r *http.Request) *core.Endpoint {
	res := r.Context().Value(gatewayCtxKeyEndpoint)
	if res == nil {
		return nil
	}
	cast, ok := res.(*core.Endpoint)
	if !ok {
		return nil
	}

	return cast
}

func ExtractContextURLParams(r *http.Request) []string {
	res := r.Context().Value(gatewayCtxKeyURLParams)
	cast, ok := res.([]string)
	if !ok {
		return nil
	}

	return cast
}

func InjectContextEndpoint(r *http.Request, contextValue *core.Endpoint) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), gatewayCtxKeyEndpoint, contextValue))
}

func InjectContextURLParams(r *http.Request, contextValue []string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), gatewayCtxKeyURLParams, contextValue))
}
