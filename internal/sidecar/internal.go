package sidecar

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type internalServer struct {
	mux    *http.ServeMux
	server *http.Server
}

func newInternalServer() *internalServer {
	slog.Info("starting internal111 server")
	mux := http.NewServeMux()

	s := &internalServer{
		mux: mux,
		server: &http.Server{
			Addr:    "127.0.0.1:8889",
			Handler: mux,
		},
	}
	s.mux.HandleFunc("/logs", s.handleLogs)

	return s
}

func (s *internalServer) handleLogs(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	defer r.Body.Close()

	fmt.Println(string(b))

	w.WriteHeader(http.StatusOK)
}
