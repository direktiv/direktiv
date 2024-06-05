package main

import (
	"net/http"
	"os"
	"strings"

	"github.com/direktiv/direktiv/cmd/cmdserver"
	"github.com/direktiv/direktiv/cmd/dinit"
	"github.com/direktiv/direktiv/cmd/sidecar"
	"github.com/direktiv/direktiv/cmd/tsengine"
	_ "github.com/direktiv/direktiv/pkg/gateway2/plugins/auth"
	_ "github.com/direktiv/direktiv/pkg/gateway2/plugins/inbound"
	_ "github.com/direktiv/direktiv/pkg/gateway2/plugins/target"
	"github.com/direktiv/direktiv/pkg/middlewares"
)

const (
	//nolint:gosec
	apiTokenHeader = "direktiv-token"
)

type apikeyHandler struct {
	next http.Handler
	key  string
}

func (h *apikeyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/api/") {
		h.next.ServeHTTP(w, r)

		return
	}
	if strings.HasSuffix(r.URL.Path, "/v2/status") {
		h.next.ServeHTTP(w, r)

		return
	}

	if r.Header.Get(apiTokenHeader) != h.key {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	h.next.ServeHTTP(w, r)
}

func main() {

	if key := os.Getenv("DIREKTIV_API_KEY"); key != "" {
		middlewares.RegisterHTTPMiddleware(func(h http.Handler) http.Handler {
			return &apikeyHandler{
				next: h,
				key:  key,
			}
		})
	}

	switch os.Getenv("DIREKTIV_APP") {
	case "sidecar":
		sidecar.RunApplication()
	case "init":
		dinit.RunApplication()
	case "tsengine":
		tsengine.RunApplication()
	case "cmdserver":
		cmdserver.RunApplication()
	default:
		// default to flow app.
		runApplication()
	}
}
