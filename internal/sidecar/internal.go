package sidecar

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/direktiv/direktiv/internal/telemetry"
)

type internalServer struct {
	mux    *http.ServeMux
	server *http.Server

	rm *requestMap
}

func newInternalServer(rm *requestMap) *internalServer {
	slog.Info("starting internal server")
	mux := http.NewServeMux()

	s := &internalServer{
		mux: mux,
		server: &http.Server{
			Addr:    "127.0.0.1:8889",
			Handler: mux,
		},
		rm: rm,
	}
	s.mux.HandleFunc("/log", s.handleLogs)

	return s
}

// legacy
func (s *internalServer) handleLogs(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	defer r.Body.Close()

	lo := s.rm.Get(r.URL.Query().Get("aid"))
	ctx := telemetry.LogInitCtx(r.Context(), lo)
	telemetry.LogInstance(ctx, telemetry.LogLevelInfo, string(b))

	w.WriteHeader(http.StatusOK)
}
