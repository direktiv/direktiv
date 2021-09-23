package api

import (
	"net/http"

	"github.com/vorteil/direktiv/pkg/flow/grpc"
	"github.com/vorteil/direktiv/pkg/util"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

// Server struct for API server
type Server struct {
	router     *mux.Router
	srv        *http.Server
	flowClient grpc.FlowClient

	config *util.Config

	//handlers
	functionHandler *functionHandler
	flowHandler     *flowHandler

	telend func()
}

// NewServer return new API server
func NewServer(l *zap.SugaredLogger) (*Server, error) {

	logger = l
	logger.Infof("starting api server")

	r := mux.NewRouter().PathPrefix("/api").Subrouter()

	s := &Server{
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

	logger.Debug("Initializing telemetry.")
	s.telend, err = util.InitTelemetry(s.config, "direktiv", "direktiv/api")
	if err != nil {
		return nil, err
	}
	r.Use(util.TelemetryMiddleware)

	s.functionHandler, err = newFunctionHandler(logger,
		r.PathPrefix("/functions").Subrouter(), s.config.FunctionsService)
	if err != nil {
		logger.Error("can not get functions handler: %v", err)
		s.telend()
		return nil, err
	}

	s.flowHandler, err = newFlowHandler(logger, r, s.config)
	if err != nil {
		logger.Error("can not get flow handler: %v", err)
		s.telend()
		return nil, err
	}

	s.prepareHelperRoutes()

	return s, nil

}

// Start starts API server
func (s *Server) Start() error {
	defer s.telend()
	logger.Infof("start listening")
	return s.srv.ListenAndServe()
}

func (s *Server) prepareHelperRoutes() {

	// Options ..
	s.router.HandleFunc("/{path:.*}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Webhook-Allowed-Origin", "eventgrid.azure.net")
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodOptions).Name(RN_Preflight)

	// functions ..
	// s.router.HandleFunc("/api/functions", s.functionHandler.listServices).Methods(http.MethodGet).Name(RN_ListServices)
	// // s.router.HandleFunc("/api/functions/pods/", s.handler.listPods).Methods(http.MethodPost).Name(RN_ListPods)
	// // s.router.HandleFunc("/api/functions/", s.handler.deleteServices).Methods(http.MethodDelete).Name(RN_DeleteServices)
	// s.router.HandleFunc("/api/functions", s.functionHandler.createService).Methods(http.MethodPost).Name(RN_CreateService)
	// s.router.HandleFunc("/api/functions/{serviceName}", s.handler.getService).Methods(http.MethodGet).Name(RN_GetService)
	// s.router.HandleFunc("/api/functions/{serviceName}", s.handler.updateService).Methods(http.MethodPost).Name(RN_UpdateService)
	// s.router.HandleFunc("/api/functions/{serviceName}", s.handler.updateServiceTraffic).Methods(http.MethodPatch).Name(RN_UpdateServiceTraffic)
	// s.router.HandleFunc("/api/functions/{serviceName}", s.handler.deleteService).Methods(http.MethodDelete).Name(RN_DeleteService)
	// s.router.HandleFunc("/api/functionrevisions/{revision}", s.handler.deleteRevision).Methods(http.MethodDelete).Name(RN_DeleteRevision)

	// engine
	// s.router.HandleFunc("/api/flow", s.flowHandler.listFunctions).Methods(http.MethodGet).Name(RN_ListServices)

	// variables

	// metrics
}
