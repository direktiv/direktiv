package sidecar

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/direktiv/direktiv/pkg/flow"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/trace"
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

func waitForUserContainer() {
	ticker := time.NewTicker(250 * time.Millisecond)

	go func() {
		time.Sleep(2 * time.Minute)
		ticker.Stop()
	}()

	for range ticker.C {
		conn, _ := net.DialTimeout("tcp", "localhost:8080", time.Second)
		if conn != nil {
			log.Debug("user container connected")
			_ = conn.Close()
			return
		}
	}

	panic("user container did not start in time")
}

// Start starts the network server for the sidecar.
func (srv *NetworkServer) Start() {
	waitForUserContainer()

	srv.router = mux.NewRouter()

	srv.router.Use(util.TelemetryMiddleware)

	srv.router.HandleFunc("/", srv.functions)

	srv.router.HandleFunc("/cancel", srv.cancel)

	port := "8890"
	if os.Getenv("DIREKITV_ENABLE_DOCKER") == "true" {
		port = "80"
	}
	srv.server.Addr = "0.0.0.0:" + port
	srv.server.Handler = srv.router

	srv.stopper = make(chan *time.Time, 1)

	srv.end = threads.Register(srv.stopper)

	log.Debug("Network-facing server thread registered.")

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
	traceID := r.Header.Get(flow.DirektivTraceIDHeader)
	spanID := r.Header.Get(flow.DirektivSpanIDHeader)
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID([]byte(traceID)),
		SpanID:  trace.SpanID([]byte(spanID)),
	})
	tctx := trace.ContextWithSpanContext(r.Context(), spanContext)
	r = r.WithContext(tctx)

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

	waiting := true
	for waiting {
		select {
		case srv.local.queue <- req:
			waiting = false
			log.Debugf("Request '%s' queued.", id)
		case <-time.After(time.Second * 30):
			log.Warnf("Request '%s' is starving!", id)
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
