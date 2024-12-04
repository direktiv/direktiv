package main

import (
	"github.com/direktiv/direktiv/pkg/cmdserver"
	"net/http"
	"os"
	"strings"

	_ "github.com/direktiv/direktiv/pkg/gateway/plugins/auth"
	_ "github.com/direktiv/direktiv/pkg/gateway/plugins/inbound"
	_ "github.com/direktiv/direktiv/pkg/gateway/plugins/outbound"
	_ "github.com/direktiv/direktiv/pkg/gateway/plugins/target"
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

	if strings.Contains(os.Args[0], "direktiv-cmd") {
		cmdserver.Start()

		return
	}

	runApplication()
}
