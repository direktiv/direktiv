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

		// Add new headers if needed
		actionID := req.Header.Get(core.EngineHeaderActionID)

		// remove all headers
		for header := range req.Header {
			req.Header.Del(header)
		}

		req.Header.Set(core.EngineHeaderActionID, actionID)

		// Log for debugging
		slog.Info("forwarding request to user container", slog.String("actionID", actionID))
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		fmt.Printf("!!!!!!!!!!!!!!!!!!!!Proxy error: %v\n", err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
	}

	proxy.ModifyResponse = func(resp *http.Response) error {
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

// func (s *externalServer) handleActionRequest(w http.ResponseWriter, r *http.Request) {
// 	b, err := io.ReadAll(r.Body)
// 	if err != nil {
// 		w.WriteHeader(500)
// 	}
// 	defer r.Body.Close()

// 	// do internal stuuf

// 	// post data to user container

// 	fmt.Println(string(b))

// 	w.WriteHeader(http.StatusOK)
// }
