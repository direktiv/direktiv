package main

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/direktiv/direktiv/cmd/flow"
	"github.com/direktiv/direktiv/cmd/sidecar"
	"github.com/direktiv/direktiv/pkg/refactor/middlewares"
)

const (
	//nolint:gosec
	APITokenHeader = "direktiv-token"
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

	if r.Header.Get(APITokenHeader) != h.key {
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
				key:  key,
			}
		})
	}

	switch appName {
	case "sidecar":
		sidecar.RunApplication()
	case "flow":
		flow.RunApplication()
	case "init":

		perm := 0o755

		// copies the command server to the shared directory in kubernetes
		err := os.MkdirAll("/usr/share/direktiv", fs.FileMode(perm))
		if err != nil {
			panic(err)
		}

		source, err := os.Open("/bin/direktiv-cmd")
		if err != nil {
			panic(err)
		}

		destination, err := os.Create("/usr/share/direktiv/direktiv-cmd")
		if err != nil {
			panic(err)
		}

		_, err = io.Copy(destination, source)
		if err != nil {
			panic(err)
		}

		err = os.Chmod("/usr/share/direktiv/direktiv-cmd", fs.FileMode(perm))
		if err != nil {
			panic(err)
		}

		destination.Close()
		source.Close()

	case "":
		log.Fatalf("error: empty DIREKTIV_APP environment variable.\n")
	default:
		log.Fatalf(fmt.Sprintf("error: invalid DIREKTIV_APP environment variable value, got: '%s'.\n", appName))
	}
}
