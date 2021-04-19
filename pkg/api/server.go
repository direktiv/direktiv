package api

import (
	"net/http"
	"os"
	"sync"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/pkg/direktiv"
	"github.com/vorteil/direktiv/pkg/ingress"
)

// Server ..
type Server struct {
	cfg      *Config
	direktiv ingress.DirektivIngressClient
	json     jsonpb.Marshaler
	handler  *Handler
	routes   map[string]map[string]http.HandlerFunc
	router   *mux.Router
	srv      *http.Server

	reqMapMutex sync.Mutex
	reqMap      map[*http.Request]*RequestStatus
}

func NewServer(cfg *Config) (*Server, error) {

	r := mux.NewRouter()

	s := &Server{
		cfg:    cfg,
		router: r,
		srv: &http.Server{
			Handler: r,
			Addr:    cfg.Server.Bind,
		},
		reqMapMutex: sync.Mutex{},
		reqMap:      make(map[*http.Request]*RequestStatus),
		json: jsonpb.Marshaler{
			EmitDefaults: true,
		},
	}

	s.handler = &Handler{
		s: s,
	}

	err := s.initDirektiv()
	if err != nil {
		return nil, err
	}

	s.prepareRoutes()

	return s, nil
}

// IngressClient returns client to backend
func (s *Server) IngressClient() ingress.DirektivIngressClient {
	return s.direktiv
}

// Router returns mux router
func (s *Server) Router() *mux.Router {
	return s.router
}

func (s *Server) initDirektiv() error {

	conn, err := direktiv.GetEndpointTLS(s.cfg.Ingress.Endpoint)
	if err != nil {
		return err
	}

	s.direktiv = ingress.NewDirektivIngressClient(conn)

	return nil
}

func (s *Server) prepareRoutes() {

	// init routes map
	s.routes = make(map[string]map[string]http.HandlerFunc)
	s.routes[http.MethodGet] = make(map[string]http.HandlerFunc)
	s.routes[http.MethodPost] = make(map[string]http.HandlerFunc)
	s.routes[http.MethodPut] = make(map[string]http.HandlerFunc)
	s.routes[http.MethodDelete] = make(map[string]http.HandlerFunc)
	s.routes[http.MethodPatch] = make(map[string]http.HandlerFunc)
	s.routes[http.MethodOptions] = make(map[string]http.HandlerFunc)

	// Options ..
	s.routes[http.MethodOptions]["/{path:.*}"] = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Webhook-Allowed-Origin", "eventgrid.azure.net")
		w.WriteHeader(http.StatusOK)
	})

	// Namespace ..
	s.routes[http.MethodGet]["/api/namespaces/"] = http.HandlerFunc(s.handler.Namespaces)
	s.routes[http.MethodPost]["/api/namespaces/{namespace}"] = http.HandlerFunc(s.handler.AddNamespace)
	s.routes[http.MethodDelete]["/api/namespaces/{namespace}"] = http.HandlerFunc(s.handler.DeleteNamespace)

	// Event ..
	s.routes[http.MethodPost]["/api/namespaces/{namespace}/event"] = http.HandlerFunc(s.handler.NamespaceEvent)

	// Secret ..
	s.routes[http.MethodGet]["/api/namespaces/{namespace}/secrets/"] = http.HandlerFunc(s.handler.Secrets)
	s.routes[http.MethodPost]["/api/namespaces/{namespace}/secrets/"] = http.HandlerFunc(s.handler.CreateSecret)
	s.routes[http.MethodDelete]["/api/namespaces/{namespace}/secrets/"] = http.HandlerFunc(s.handler.DeleteSecret)

	// Registry ..
	s.routes[http.MethodGet]["/api/namespaces/{namespace}/registries/"] = http.HandlerFunc(s.handler.Registries)
	s.routes[http.MethodPost]["/api/namespaces/{namespace}/registries/"] = http.HandlerFunc(s.handler.CreateRegistry)
	s.routes[http.MethodDelete]["/api/namespaces/{namespace}/registries/"] = http.HandlerFunc(s.handler.DeleteRegistry)

	// Workflow ..
	s.routes[http.MethodGet]["/api/namespaces/{namespace}/workflows/"] = http.HandlerFunc(s.handler.Workflows)
	s.routes[http.MethodGet]["/api/namespaces/{namespace}/workflows/{workflowTarget}"] = http.HandlerFunc(s.handler.GetWorkflow)
	s.routes[http.MethodPut]["/api/namespaces/{namespace}/workflows/{workflowUID}"] = http.HandlerFunc(s.handler.UpdateWorkflow)
	s.routes[http.MethodPut]["/api/namespaces/{namespace}/workflows/{workflowUID}/toggle"] = http.HandlerFunc(s.handler.ToggleWorkflow)
	s.routes[http.MethodPost]["/api/namespaces/{namespace}/workflows"] = http.HandlerFunc(s.handler.CreateWorkflow)
	s.routes[http.MethodDelete]["/api/namespaces/{namespace}/workflows/{workflowUID}"] = http.HandlerFunc(s.handler.DeleteWorkflow)
	s.routes[http.MethodGet]["/api/namespaces/{namespace}/workflows/{workflowUID}/download"] = http.HandlerFunc(s.handler.DownloadWorkflow)
	s.routes[http.MethodPost]["/api/namespaces/{namespace}/workflows/{workflowID}/execute"] = http.HandlerFunc(s.handler.ExecuteWorkflow)

	// Metrics ..
	s.routes[http.MethodGet]["/namespaces/{namespace}/workflows/{workflow}/metrics"] = http.HandlerFunc(s.handler.WorkflowMetrics)

	// Instance ..
	s.routes[http.MethodGet]["/api/instances/{namespace}"] = http.HandlerFunc(s.handler.Instances)
	s.routes[http.MethodGet]["/api/instances/{namespace}/{workflowID}/{id}"] = http.HandlerFunc(s.handler.GetInstance)
	s.routes[http.MethodDelete]["/api/instances/{namespace}/{workflowID}/{id}"] = http.HandlerFunc(s.handler.CancelInstance)
	s.routes[http.MethodGet]["/api/instances/{namespace}/{workflowID}/{id}/logs"] = http.HandlerFunc(s.handler.InstanceLogs)

}

// RegisterHandler registers all handlers
func (s *Server) RegisterHandler(path string, h http.HandlerFunc, methods ...string) {
	for _, method := range methods {
		s.routes[method][path] = h
	}
}

// Start starts the API server
func (s *Server) Start() error {

	for method, paths := range s.routes {
		for path, h := range paths {
			s.router.HandleFunc(path, h).Methods(method)
		}
	}

	log.Infof("Starting server - binding to %s\n", s.cfg.Server.Bind)

	if _, err := os.Stat(direktiv.TLSCert); !os.IsNotExist(err) {
		log.Infof("tls enabled")
		return s.srv.ListenAndServeTLS(direktiv.TLSCert, direktiv.TLSKey)
	}

	return s.srv.ListenAndServe()
}
