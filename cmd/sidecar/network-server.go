package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/direktiv/direktiv/pkg/util"
	"github.com/gorilla/mux"
)

const actionIDHeader = "Direktiv-ActionID"

// NetworkServer defines a network server object
type NetworkServer struct {
	end     func()
	local   *LocalServer
	router  *mux.Router
	server  http.Server
	stopper chan *time.Time
}

// Start starts the network server for the sidecar
func (srv *NetworkServer) Start() {

	// knative does not support startup probes, so we need to wait heer for port 8080
	for {
		conn, _ := net.DialTimeout("tcp", "localhost:8080", time.Second)
		if conn != nil {
			conn.Close()
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	srv.router = mux.NewRouter()

	srv.router.Use(util.TelemetryMiddleware)

	srv.router.HandleFunc("/", srv.functions)

	srv.server.Addr = "0.0.0.0:8890"
	srv.server.Handler = srv.router

	srv.stopper = make(chan *time.Time, 1)

	srv.end = threads.Register(srv.stopper)

	log.Debug("Network-facing server thread registered.")

	go srv.run()
	go srv.wait()

}

func (srv *NetworkServer) wait() {

	defer srv.server.Close()
	defer srv.end()

	t := <-srv.stopper

	log.Debug("Network-facing server shutting down.")

	ctx, cancel := context.WithDeadline(context.Background(), t.Add(15*time.Second))
	defer cancel()

	err := srv.server.Shutdown(ctx)
	if err != nil {
		log.Errorf("Error shutting down network-facing server: %v", err)
		Shutdown(ERROR)
		return
	}

	log.Debug("Network-facing server shut down successfully.")

}

func (srv *NetworkServer) run() {

	log.Infof("Starting network-facing HTTP server on %s.", srv.server.Addr)

	err := srv.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Errorf("Error running network-facing server: %v", err)
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
			log.Errorf("Request '%s' panicked: %v.", id, r)
			srv.local.drainRequest(req)
		} else {
			_ = req.r.Body.Close()
		}

	}(req)

	var waiting = true
	for waiting {
		select {
		case srv.local.queue <- req:
			waiting = false
			log.Debugf("Request '%s' queued.", id)
		case <-time.After(time.Second * 30):
			log.Warnf("Request '%s' is starving!", id)
			// TODO: reject request after some number of failures
		}
	}

	waiting = true
	for waiting {
		select {
		case <-req.end:
			waiting = false
			log.Debugf("Request '%s' returned.", id)
		case <-time.After(time.Minute * 5):
			log.Infof("Request '%s' hasn't returned yet.", id)
		}
	}

}
