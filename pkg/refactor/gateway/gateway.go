package gateway

import (
	"encoding/json"
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/core"
)

type RouteConfiguration struct{}

type Handler struct{}

func NewHandler() *Handler {
	gw := &Handler{}

	return gw
}

func (gw *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	prefix := "/api/v2/gateway"
	if r.Method == "GET" && r.URL.Path == prefix {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		payLoad := struct {
			Data any `json:"data"`
		}{
			Data: "hi",
		}
		_ = json.NewEncoder(w).Encode(payLoad)

		return
	}
}

func (gw *Handler) SetEndpoints(list []*core.Endpoint) {
}

func (gw *Handler) ListEndpoints() []*core.Endpoint {
	return nil
}
