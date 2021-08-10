package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/vorteil/direktiv/pkg/api"
	_ "github.com/vorteil/direktiv/pkg/util"
)

// Envrioment Value Name for Manual apikey
const API_KEY_ENV = "DIREKTIV_API_KEY"

// apiKey to be used for simple apikey auth
var apiKey = ""

func main() {

	cfg, err := api.Configure()
	if err != nil {
		log.Fatalf(err.Error())
	}

	s, err := api.NewServer(cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Enabled simple apikey auth if env is set
	apiKey = os.Getenv(API_KEY_ENV)
	if len(apiKey) > 0 {
		fmt.Println("apiKey auth mode enabled")
		s.Router().Use(APIAuthMiddleware)
	}

	err = s.Start()
	if err != nil {
		log.Fatal(err.Error())
	}

}

// API Auth Middleware
func APIAuthMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow preflight Options
		if r.Method == "OPTIONS" {
			h.ServeHTTP(w, r)
			return
		}

		// Get Auth token
		reqToken := r.Header.Get("Authorization")
		splitToken := strings.Split(reqToken, "apikey")
		if len(splitToken) != 2 {
			w.WriteHeader(401)
			/* #nosec */
			_, _ = w.Write([]byte("apikey Token is malformed"))
			return
		}

		reqToken = strings.TrimSpace(splitToken[1])

		if reqToken != apiKey {
			w.WriteHeader(401)
			return
		}

		h.ServeHTTP(w, r)
	})
}
