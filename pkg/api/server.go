package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/middlewares"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

// Server struct for API server.
type Server struct {
	logger *zap.SugaredLogger
	router *mux.Router
	srv    *http.Server

	config *core.Config

	// handlers
	flowHandler *flowHandler

	telend func()
}

// GetRouter is a getter for s.router.
func (s *Server) GetRouter() *mux.Router {
	return s.router
}

// NewServer return new API server.
func NewServer(l *zap.SugaredLogger, config *core.Config) (*Server, error) {
	logger = l
	logger.Infof("starting api server")

	baseRouter := mux.NewRouter()
	r := baseRouter.PathPrefix("/api").Subrouter()

	s := &Server{
		config: config,
		logger: l,
		router: r,
		srv: &http.Server{
			Handler:           baseRouter,
			Addr:              fmt.Sprintf(":%v", config.ApiV1Port),
			ReadHeaderTimeout: time.Second * 60,
		},
	}

	// cast to gorilla mux type
	var gorillaMiddlewares []mux.MiddlewareFunc
	for i := range middlewares.GetMiddlewares() {
		gorillaMiddlewares = append(gorillaMiddlewares, mux.MiddlewareFunc(middlewares.GetMiddlewares()[i]))
	}
	gorillaMiddlewares = append(gorillaMiddlewares, s.logMiddleware)

	r.Use(gorillaMiddlewares...)

	var err error

	s.flowHandler, err = newFlowHandler(logger, baseRouter, r, s.config)
	if err != nil {
		logger.Error("can not get flow handler: %v", err)
		s.telend()
		return nil, err
	}

	logger.Debug("adding options routes")
	s.prepareHelperRoutes()

	return s, nil
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
