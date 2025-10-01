package sidecar

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/direktiv/direktiv/internal/core"
)

type externalServer struct {
	server *http.Server
}

func newExternalServer() *externalServer {

	// we can ignore the error here
	addr, _ := url.Parse("http://localhost:8080")
	proxy := httputil.NewSingleHostReverseProxy(addr)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		if req.URL.Path == "/up" {
			// status request to avoid retry
			// this makes the proxy fail
			req.URL = nil
			return
		}

		// add action header
		actionID := req.Header.Get(core.EngineHeaderActionID)

		// remove all headers
		for header := range req.Header {
			req.Header.Del(header)
		}

		req.Header.Set(core.EngineHeaderActionID, actionID)

		// TODO: create temp directory
		req.Header.Set(core.EngineHeaderTempDir, "/tmp")

		// Log for debugging
		slog.Info("forwarding request to user container", slog.String("actionID", actionID))
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		if r.URL == nil {
			w.WriteHeader(http.StatusOK)
			return
		}
		fmt.Println("ERROR HANDLING!!!")
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
	}

	proxy.ModifyResponse = func(resp *http.Response) error {
		// if it is not ok, we return 502 to trigger the retry
		if resp.StatusCode != http.StatusOK {
			resp.StatusCode = 502
		}
		return nil
	}

	slog.Info("starting external proxy")

	s := &externalServer{
		server: &http.Server{
			Addr:    "0.0.0.0:8890",
			Handler: proxy,
		},
	}

	return s
}
