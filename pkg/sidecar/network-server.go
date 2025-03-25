package sidecar

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

const (
	actionIDHeader = "Direktiv-ActionID"
	IteratorHeader = "Direktiv-Iterator"
)

// NetworkServer defines a network server object.
type NetworkServer struct {
	end     func()
	local   *LocalServer
	router  *mux.Router
	server  http.Server
	stopper chan *time.Time
}

// waitForUserContainer waits for a user-defined container to start and become
// available by attempting to connect to localhost:8080. It waits for up to 2 minutes
// before terminating the wait and panicking.
func waitForUserContainer() {
	slog.Debug("waiting for user container to become available")

	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	timeout := time.After(2 * time.Minute)

	for {
		select {
		case <-ticker.C:
			conn, err := net.DialTimeout("tcp", "localhost:8080", time.Second)
			if err != nil {
				slog.Debug("failed to connect to user container", "error", err)
				continue
			}
			slog.Debug("user container is now available", "address", "localhost:8080")
			_ = conn.Close()

			return
		case <-timeout:
			panic("user container did not start in time. timeout waiting for connection to localhost:8080")
		}
	}
}

// Start starts the network server for the sidecar.
func (srv *NetworkServer) Start() {
	waitForUserContainer()

	srv.router = mux.NewRouter()
	srv.router.HandleFunc("/", srv.functions)
	srv.router.HandleFunc("/cancel", srv.cancel)

	port := "8890"
	srv.server.Addr = "0.0.0.0:" + port
	srv.server.Handler = srv.router

	srv.stopper = make(chan *time.Time, 1)

	srv.end = threads.Register(srv.stopper)

	slog.Debug("network-facing server thread registered")

	go srv.run()
	go srv.wait()
}

func (srv *NetworkServer) cancel(w http.ResponseWriter, r *http.Request) {
	srv.local.cancelActiveRequest(context.Background(),
		r.Header.Get(actionIDHeader))
}

func (srv *NetworkServer) wait() {
	defer srv.server.Close()
	defer srv.end()

	t := <-srv.stopper

	slog.Debug("network-facing server shutting down")

	ctx, cancel := context.WithDeadline(context.Background(), t.Add(15*time.Second))
	defer cancel()

	err := srv.server.Shutdown(ctx)
	if err != nil {
		slog.Error("error shutting down network-facing server", "error", err)
		Shutdown(ERROR)

		return
	}

	slog.Debug("network-facing server shut down successfully")
}

func (srv *NetworkServer) run() {
	slog.Info("starting network-facing HTTP server", "addr", srv.server.Addr)

	err := srv.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("error running network-facing server", "error", err)
		Shutdown(ERROR)

		return
	}
}

func (srv *NetworkServer) functions(w http.ResponseWriter, r *http.Request) {
	req := &inboundRequest{
		w:   w,
		r:   r,
		end: make(chan bool),
	}

	id := r.Header.Get(actionIDHeader)

	defer func(req *inboundRequest) {
		r := recover()
		if r != nil {
			slog.Error("request panicked", "action-id", id, "request", r)
			srv.local.drainRequest(req)
		} else {
			_ = req.r.Body.Close()
		}
	}(req)

	waiting := true
	for waiting {
		select {
		case srv.local.queue <- req:
			waiting = false
			slog.Debug("request queued", "action_id", id)
		case <-time.After(time.Second * 30):
			slog.Warn("Request is starving!", "action_id", id)
		}
	}

	waiting = true
	for waiting {
		select {
		case <-req.end:
			waiting = false
			slog.Debug("request returned", "action-id", id)
		case <-time.After(time.Minute * 5):
			slog.Info("request hasn't returned yet", "action-id", id)
		}
	}
}
