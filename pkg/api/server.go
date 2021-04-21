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

	wfTemplateDirsPaths map[string]string
	wfTemplateDirs      []string

	actionTemplateDirsPaths map[string]string
	actionTemplateDirs      []string
}

// NewServer returns new API server
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

	err = s.initTemplates()
	if err != nil {
		return nil, err
	}

	s.prepareRoutes()

	return s, nil
}

func (s *Server) initTemplates() error {

	err := s.initWorkflowTemplates()
	if err != nil {
		return err
	}

	err = s.initActionTemplates()
	if err != nil {
		return err
	}

	return nil
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
	s.routes[http.MethodGet]["/api/namespaces/"] = http.HandlerFunc(s.handler.namespaces)
	s.routes[http.MethodPost]["/api/namespaces/{namespace}"] = http.HandlerFunc(s.handler.addNamespace)
	s.routes[http.MethodDelete]["/api/namespaces/{namespace}"] = http.HandlerFunc(s.handler.deleteNamespace)

	// Event ..
	s.routes[http.MethodPost]["/api/namespaces/{namespace}/event"] = http.HandlerFunc(s.handler.namespaceEvent)

	// Secret ..
	s.routes[http.MethodGet]["/api/namespaces/{namespace}/secrets/"] = http.HandlerFunc(s.handler.secrets)
	s.routes[http.MethodPost]["/api/namespaces/{namespace}/secrets/"] = http.HandlerFunc(s.handler.createSecret)
	s.routes[http.MethodDelete]["/api/namespaces/{namespace}/secrets/"] = http.HandlerFunc(s.handler.deleteSecret)

	// Registry ..
	s.routes[http.MethodGet]["/api/namespaces/{namespace}/registries/"] = http.HandlerFunc(s.handler.registries)
	s.routes[http.MethodPost]["/api/namespaces/{namespace}/registries/"] = http.HandlerFunc(s.handler.createRegistry)
	s.routes[http.MethodDelete]["/api/namespaces/{namespace}/registries/"] = http.HandlerFunc(s.handler.deleteRegistry)

	// Metrics ..
	s.routes[http.MethodGet]["/api/namespaces/{namespace}/workflows/{workflow}/metrics"] = http.HandlerFunc(s.handler.workflowMetrics)

	// Workflow ..
	s.routes[http.MethodGet]["/api/namespaces/{namespace}/workflows/"] = http.HandlerFunc(s.handler.workflows)
	s.routes[http.MethodGet]["/api/namespaces/{namespace}/workflows/{workflowTarget}"] = http.HandlerFunc(s.handler.getWorkflow)
	s.routes[http.MethodPut]["/api/namespaces/{namespace}/workflows/{workflowTarget}"] = http.HandlerFunc(s.handler.updateWorkflow)
	s.routes[http.MethodPut]["/api/namespaces/{namespace}/workflows/{workflowTarget}/toggle"] = http.HandlerFunc(s.handler.toggleWorkflow)
	s.routes[http.MethodPost]["/api/namespaces/{namespace}/workflows"] = http.HandlerFunc(s.handler.createWorkflow)
	s.routes[http.MethodDelete]["/api/namespaces/{namespace}/workflows/{workflowTarget}"] = http.HandlerFunc(s.handler.deleteWorkflow)
	s.routes[http.MethodGet]["/api/namespaces/{namespace}/workflows/{workflowTarget}/download"] = http.HandlerFunc(s.handler.downloadWorkflow)
	s.routes[http.MethodPost]["/api/namespaces/{namespace}/workflows/{workflowTarget}/execute"] = http.HandlerFunc(s.handler.executeWorkflow)
	s.routes[http.MethodGet]["/api/namespaces/{namespace}/workflows/{workflowTarget}/instances/"] = http.HandlerFunc(s.handler.workflowInstances)

	// Instance ..
	s.routes[http.MethodGet]["/api/instances/{namespace}"] = http.HandlerFunc(s.handler.instances)
	s.routes[http.MethodGet]["/api/instances/{namespace}/{workflowTarget}/{id}"] = http.HandlerFunc(s.handler.getInstance)
	s.routes[http.MethodDelete]["/api/instances/{namespace}/{workflowTarget}/{id}"] = http.HandlerFunc(s.handler.cancelInstance)
	s.routes[http.MethodGet]["/api/instances/{namespace}/{workflowTarget}/{id}/logs"] = http.HandlerFunc(s.handler.instanceLogs)

	// Templates ..
	s.routes[http.MethodGet]["/api/action-templates/"] = http.HandlerFunc(s.handler.actionTemplateFolders)
	s.routes[http.MethodGet]["/api/action-templates/{folder}/"] = http.HandlerFunc(s.handler.actionTemplates)
	s.routes[http.MethodGet]["/api/action-templates/{folder}/{template}"] = http.HandlerFunc(s.handler.actionTemplate)

	s.routes[http.MethodGet]["/api/workflow-templates/"] = http.HandlerFunc(s.handler.workflowTemplateFolders)
	s.routes[http.MethodGet]["/api/workflow-templates/{folder}/"] = http.HandlerFunc(s.handler.workflowTemplates)
	s.routes[http.MethodGet]["/api/workflow-templates/{folder}/{template}"] = http.HandlerFunc(s.handler.workflowTemplate)

	// jq Playground ...
	s.routes[http.MethodPost]["/api/jq-playground"] = http.HandlerFunc(s.handler.jqPlayground)

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
