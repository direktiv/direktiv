package api

import (
	"net/http"
	"time"

	"github.com/direktiv/direktiv/pkg/util"
	"github.com/direktiv/direktiv/pkg/version"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

var logger *zap.SugaredLogger

// Server struct for API server.
type Server struct {
	logger *zap.SugaredLogger
	router *mux.Router
	srv    *http.Server

	config *util.Config

	// handlers
	flowHandler *flowHandler

	telend func()
}

// GetRouter is a getter for s.router.
func (s *Server) GetRouter() *mux.Router {
	return s.router
}

type mw struct {
	h http.Handler
}

func (mw *mw) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger.Infof("request received: %v %v", r.Method, r.URL.String())
	mw.h.ServeHTTP(w, r)
}

// NewServer return new API server.
func NewServer(l *zap.SugaredLogger) (*Server, error) {
	logger = l
	logger.Infof("starting api server")

	r := mux.NewRouter().PathPrefix("/api").Subrouter()

	s := &Server{
		logger: l,
		router: r,
		srv: &http.Server{
			Handler:           r,
			Addr:              ":1644",
			ReadHeaderTimeout: time.Second * 60,
		},
	}

	// swagger:operation GET /api/version Other version
	// ---
	// description: |
	//   Returns version information for servers in the cluster.
	// summary: Returns version information for servers in the cluster.
	// responses:
	//   '200':
	//     "description": "version query was successful"
	r.HandleFunc("/version", s.version).Name(RN_Version).Methods(http.MethodGet)

	// read config
	conf, err := util.ReadConfig("/etc/direktiv/flow-config.yaml")
	if err != nil {
		return nil, err
	}
	s.config = conf

	r.Use(func(h http.Handler) http.Handler {
		return &mw{h: h}
	})

	logger.Debug("Initializing telemetry.")
	s.telend, err = util.InitTelemetry(s.config, "direktiv/api", "direktiv")
	if err != nil {
		return nil, err
	}
	r.Use(util.TelemetryMiddleware, s.logMiddleware)

	s.flowHandler, err = newFlowHandler(logger, r, s.config)
	if err != nil {
		logger.Error("can not get flow handler: %v", err)
		s.telend()
		return nil, err
	}

	logger.Debug("adding options routes")
	s.prepareHelperRoutes()

	return s, nil
}

func (s *Server) version(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	m := make(map[string]string)
	m["api"] = version.Version

	flowResp, _ := s.flowHandler.client.Build(ctx, &emptypb.Empty{})
	if flowResp != nil {
		m["flow"] = flowResp.GetBuild()
	}

	funcsResp, _ := s.flowHandler.functionsClient.Build(ctx, &emptypb.Empty{})
	if flowResp != nil {
		m["functions"] = funcsResp.GetBuild()
	}

	respondJSON(w, m, nil)
}

// Start starts API server.
func (s *Server) Start() error {
	defer s.telend()
	logger.Infof("start listening")
	return s.srv.ListenAndServe()
}

func (s *Server) prepareHelperRoutes() {
	// Options ..
	s.router.HandleFunc("/{path:.*}", func(w http.ResponseWriter, r *http.Request) {
		// https://github.com/cloudevents/spec/blob/v1.0/http-webhook.md#4-abuse-protection
		w.Header().Add("WebHook-Allowed-Rate", "120")
		w.Header().Add("Webhook-Allowed-Origin", "eventgrid.azure.net")
		w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Add("Pragma", "no-cache")
		w.Header().Add("Expires", "0")
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodOptions).Name(RN_Preflight)
}
