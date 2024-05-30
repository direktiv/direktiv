package middlewares

import "net/http"

type Middleware func(http.Handler) http.Handler

var registry []Middleware

func RegisterHTTPMiddleware(m Middleware) {
	registry = append(registry, m)
}

func GetMiddlewares() []Middleware {
	return registry
}
