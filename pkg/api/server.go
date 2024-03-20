package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/middlewares"
	"github.com/gorilla/mux"
)

// Server struct for API server.
type Server struct {
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
func NewServer(ctx context.Context, config *core.Config) (*Server, error) {
	slog.Debug("starting api server")

	baseRouter := mux.NewRouter()
	r := baseRouter.PathPrefix("/api").Subrouter()

	s := &Server{
		config: config,
		router: r,
		srv: &http.Server{
			Handler:           baseRouter,
			Addr:              fmt.Sprintf(":%v", config.ApiV1Port),
			ReadHeaderTimeout: time.Second * 60,
		},
	}

	// cast to gorilla mux type
	var gorillaMiddlewares []mux.MiddlewareFunc
	slog.Debug("Adding middlewares")
	for i := range middlewares.GetMiddlewares() {
		gorillaMiddlewares = append(gorillaMiddlewares, mux.MiddlewareFunc(middlewares.GetMiddlewares()[i]))
	}
	gorillaMiddlewares = append(gorillaMiddlewares, s.logMiddleware)

	r.Use(gorillaMiddlewares...)
	slog.Info("Middlewares where added")
	var err error

	s.flowHandler, err = newFlowHandler(baseRouter, r, s.config)
	if err != nil {
		slog.Error("can not get flow handler", "error", err)
		s.telend()
		return nil, err
	}

	slog.Debug("adding options routes")
	s.prepareHelperRoutes()
	slog.Info("API server started")
	return s, nil
}

// Start starts API server.
func (s *Server) Start() error {
	defer s.telend()
	slog.Debug("server starts listening")
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
