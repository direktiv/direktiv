package server

import (
	"net/http"
)

type RouteManagerAPI struct {
}

func NewRouteManagerAPI(eeConfig *Config) (*RouteManagerAPI, error) {

	return &RouteManagerAPI{}, nil
}

func (rm *RouteManagerAPI) IsAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {

		// headerValue := r.Header.Get(APITokenHeader)
		// rctx := chi.RouteContext(r.Context())

		// only check if an API key has been set
		// not checking ui route at all
		// if !strings.HasPrefix(rctx.RoutePattern(), "/ui") &&
		// 	// APIKey != "" && headerValue != APIKey {
		// 	rw.WriteHeader(http.StatusUnauthorized)
		// 	return
		// }

		next.ServeHTTP(rw, r)
	})
}

// func (rm *RouteManagerAPI) AddExtraRoutes(r *gin.Engine) error {
// 	return nil
// }
