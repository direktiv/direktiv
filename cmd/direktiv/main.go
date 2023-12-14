package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/direktiv/direktiv/cmd/flow"
	"github.com/direktiv/direktiv/cmd/sidecar"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/middlewares"
)

const (
	//nolint:gosec
	// G101 -- This is a false positive.
	APITokenHeader = "direktiv-token"
)

type apikeyHandler struct {
	next http.Handler
	key  []byte
}

func (h *apikeyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/api/") {
		h.next.ServeHTTP(w, r)

		return
	}

	token := r.Header.Get(APITokenHeader)

	expectedHMAC := core.ComputeHMAC(token, h.key)

	// Compare the expected HMAC with the actual HMAC in a constant time manner.
	if !core.CompareHMAC(expectedHMAC, []byte(token)) {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}
	h.next.ServeHTTP(w, r)
}

func main() {
	appName := os.Getenv("DIREKTIV_APP")

	if key := os.Getenv("DIREKTIV_API_KEY"); key != "" {
		middlewares.RegisterHTTPMiddleware(func(h http.Handler) http.Handler {
			return &apikeyHandler{
				next: h,
				key:  []byte(key),
			}
		})
	}

	switch appName {
	case "sidecar":
		sidecar.RunApplication()
	case "flow":
		flow.RunApplication()
	case "":
		log.Fatalf("error: empty DIREKTIV_APP environment variable.\n")
	default:
		log.Fatalf(fmt.Sprintf("error: invalid DIREKTIV_APP environment variable value, got: '%s'.\n", appName))
	}
}
