package api

import (
	"net/http"

	"github.com/vorteil/direktiv/pkg/flow/grpc"
	"github.com/vorteil/direktiv/pkg/util"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

var log *zap.SugaredLogger

// Server struct for API server
type Server struct {
	logger *zap.SugaredLogger

	router     *mux.Router
	srv        *http.Server
	flowClient grpc.FlowClient

	config *util.Config

	//handlers
	functionHandler *functionHandler
	flowHandler     *flowHandler
}

// NewServer return new API server
func NewServer(logger *zap.SugaredLogger) (*Server, error) {

	logger.Infof("starting api server")

	log = logger

	r := mux.NewRouter()

	s := &Server{
		logger: logger,
		router: r,
		srv: &http.Server{
			Handler: r,
			Addr:    ":8080",
		},
	}

	// read config
	conf, err := util.ReadConfig("/etc/direktiv/flow-config.yaml")
	if err != nil {
		return nil, err
	}
	s.config = conf

	s.functionHandler, err = newFunctionHandler(logger, s.config.FunctionsService)
	if err != nil {
		logger.Error("can not get functions handler: %v", err)
		return nil, err
	}

	s.flowHandler, err = newFlowHandler(logger, r, s.config.FlowService)
	if err != nil {
		logger.Error("can not get flow handler: %v", err)
		return nil, err
	}

	s.prepareRoutes()

	return s, nil

}

// Start starts API server
func (s *Server) Start() error {
	s.logger.Infof("start listening")
	return s.srv.ListenAndServe()
}

func (s *Server) prepareRoutes() {

	// Options ..
	s.router.HandleFunc("/{path:.*}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Webhook-Allowed-Origin", "eventgrid.azure.net")
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodOptions).Name(RN_Preflight)

	// functions
	s.router.HandleFunc("/api/functions", s.functionHandler.listFunctions).Methods(http.MethodGet).Name(RN_ListServices)

	// engine
	s.router.HandleFunc("/api/flow", s.flowHandler.listFunctions).Methods(http.MethodGet).Name(RN_ListServices)

	// variables

	// metrics
}
