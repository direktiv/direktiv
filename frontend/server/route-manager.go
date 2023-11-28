package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
)

type RouteManagerAPI struct {
	host   string
	config *Config
}

func NewRouteManagerAPI(config *Config) (*RouteManagerAPI, error) {
	return &RouteManagerAPI{
		host:   config.Server.Backend,
		config: config,
	}, nil
}

func (rm *RouteManagerAPI) AddExtraRoutes(r *chi.Mux) error {

	r.Handle("/api/*",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiPath := chi.URLParam(r, "*")
			fp := fmt.Sprintf("%s/api/%s", rm.host, apiPath)
			ReverseProxy(r, w, fp)
		}))
	return nil
}
