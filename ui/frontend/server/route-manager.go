package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
)

type RouteManagerAPI struct {
	host   string
	hostv2 string
	config *Config
}

func NewRouteManagerAPI(config *Config) (*RouteManagerAPI, error) {
	return &RouteManagerAPI{
		host:   config.Server.Backend,
		config: config,
		hostv2: config.Server.BackendV2,
	}, nil
}

func (rm *RouteManagerAPI) AddExtraRoutes(r *chi.Mux) error {
	r.Handle("/api/*",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiPath := chi.URLParam(r, "*")
			fp := fmt.Sprintf("%s/api/%s", rm.host, apiPath)
			ReverseProxy(r, w, fp)
		}))

	r.With(rm.IsAuthenticated).Handle("/gw/*",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiPath := chi.URLParam(r, "*")
			fp := fmt.Sprintf("%s/gw/%s", rm.hostv2, apiPath)
			ReverseProxy(r, w, fp)
		}))

	r.With(rm.IsAuthenticated).Handle("/ns/*",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiPath := chi.URLParam(r, "*")
			fp := fmt.Sprintf("%s/ns/%s", rm.hostv2, apiPath)
			ReverseProxy(r, w, fp)
		}))
	return nil
}
