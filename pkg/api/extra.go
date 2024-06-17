package api

import (
	"github.com/go-chi/chi/v5"
)

var extraRoutes []func(*chi.Mux)

func RegisterExtraRoute(fn func(*chi.Mux)) {
	extraRoutes = append(extraRoutes, fn)
}

func GetExtraRoutes() []func(*chi.Mux) {
	return extraRoutes
}
