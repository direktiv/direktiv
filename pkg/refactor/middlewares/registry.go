package middlewares

import "net/http"

type middleware func(http.Handler) http.Handler

var registry []middleware

func RegisterHTTPMiddleware(m middleware) {
	registry = append(registry, m)
}

func GetMiddlewares() []middleware {
	return registry
}
