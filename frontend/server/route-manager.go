package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/spf13/viper"
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

	r.With(rm.IsAuthenticated).Handle("/api/*",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiPath := chi.URLParam(r, "*")
			fp := fmt.Sprintf("%s/api/%s", rm.host, apiPath)
			ReverseProxy(r, w, fp)
		}))
	return nil
}

func (rm *RouteManagerAPI) IsAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {

		headerValue := r.Header.Get(APITokenHeader)
		apikey := viper.GetString("server.apikey")

		// only check if an API key has been set, not checking ui route at all
		if !strings.HasPrefix(r.URL.RequestURI(), "/ui") &&
			apikey != "" && headerValue != apikey {
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(rw, r)
	})
}
