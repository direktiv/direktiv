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
	"google.golang.org/grpc/resolver"
)

// Server ..
type Server struct {
	cfg      *Config
	direktiv ingress.DirektivIngressClient
	json     jsonpb.Marshaler
	handler  *Handler
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

	conn, err := direktiv.GetEndpointTLS(s.cfg.Ingress.Endpoint, true)
	if err != nil {
		return err
	}

	log.Infof("connecting to %s", s.cfg.Ingress.Endpoint)

	s.direktiv = ingress.NewDirektivIngressClient(conn)

	return nil
}

func (s *Server) prepareRoutes() {

	// Options ..
	s.Router().HandleFunc("/{path:.*}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Webhook-Allowed-Origin", "eventgrid.azure.net")
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodOptions).Name("Preflight")

	// Namespace ..
	s.Router().HandleFunc("/api/namespaces/", s.handler.namespaces).Methods(http.MethodGet).Name("ListNamespaces")
	s.Router().HandleFunc("/api/namespaces/{namespace}", s.handler.addNamespace).Methods(http.MethodPost).Name("AddNamespace")
	s.Router().HandleFunc("/api/namespaces/{namespace}", s.handler.deleteNamespace).Methods(http.MethodDelete).Name("DeleteNamespace")

	// Event ..
	s.Router().HandleFunc("/api/namespaces/{namespace}/event", s.handler.namespaceEvent).Methods(http.MethodPost).Name("NamespaceEvent")

	// Secret ..
	s.Router().HandleFunc("/api/namespaces/{namespace}/secrets/", s.handler.secrets).Methods(http.MethodGet).Name("ListSecrets")
	s.Router().HandleFunc("/api/namespaces/{namespace}/secrets/", s.handler.createSecret).Methods(http.MethodPost).Name("CreateSecret")
	s.Router().HandleFunc("/api/namespaces/{namespace}/secrets/", s.handler.deleteSecret).Methods(http.MethodDelete).Name("DeleteSecret")

	// Registry ..
	s.Router().HandleFunc("/api/namespaces/{namespace}/registries/", s.handler.registries).Methods(http.MethodGet).Name("ListRegistries")
	s.Router().HandleFunc("/api/namespaces/{namespace}/registries/", s.handler.createRegistry).Methods(http.MethodPost).Name("CreateRegistry")
	s.Router().HandleFunc("/api/namespaces/{namespace}/registries/", s.handler.deleteRegistry).Methods(http.MethodDelete).Name("DeleteRegistry")

	// Metrics ..
	s.Router().HandleFunc("/api/namespaces/{namespace}/workflows/{workflow}/metrics", s.handler.workflowMetrics).Methods(http.MethodGet).Name("GetWorkflowMetrics")

	// Workflow ..
	s.Router().HandleFunc("/api/namespaces/{namespace}/workflows/", s.handler.workflows).Methods(http.MethodGet).Name("ListWorkflows")
	s.Router().HandleFunc("/api/namespaces/{namespace}/workflows/{workflowTarget}", s.handler.getWorkflow).Methods(http.MethodGet).Name("GetWorkflow")
	s.Router().HandleFunc("/api/namespaces/{namespace}/workflows/{workflowTarget}", s.handler.updateWorkflow).Methods(http.MethodPut).Name("UpdateWorkflow")
	s.Router().HandleFunc("/api/namespaces/{namespace}/workflows/{workflowTarget}/toggle", s.handler.toggleWorkflow).Methods(http.MethodPut).Name("ToggleWorkflow")
	s.Router().HandleFunc("/api/namespaces/{namespace}/workflows", s.handler.createWorkflow).Methods(http.MethodPost).Name("CreateWorkflow")
	s.Router().HandleFunc("/api/namespaces/{namespace}/workflows/{workflowTarget}", s.handler.deleteWorkflow).Methods(http.MethodDelete).Name("DeleteWorkflow")
	s.Router().HandleFunc("/api/namespaces/{namespace}/workflows/{workflowTarget}/download", s.handler.downloadWorkflow).Methods(http.MethodGet).Name("DownloadWorkflow")
	s.Router().HandleFunc("/api/namespaces/{namespace}/workflows/{workflowTarget}/execute", s.handler.executeWorkflow).Methods(http.MethodPost).Name("ExecuteWorkflow")
	s.Router().HandleFunc("/api/namespaces/{namespace}/workflows/{workflowTarget}/instances/", s.handler.workflowInstances).Methods(http.MethodGet).Name("ListWorkflowInstances")

	// Instance ..
	s.Router().HandleFunc("/api/instances/{namespace}", s.handler.instances).Methods(http.MethodGet).Name("ListInstances")
	s.Router().HandleFunc("/api/instances/{namespace}/{workflowTarget}/{id}", s.handler.getInstance).Methods(http.MethodGet).Name("GetInstance")
	s.Router().HandleFunc("/api/instances/{namespace}/{workflowTarget}/{id}", s.handler.cancelInstance).Methods(http.MethodDelete).Name("CancelInstance")
	s.Router().HandleFunc("/api/instances/{namespace}/{workflowTarget}/{id}/logs", s.handler.instanceLogs).Methods(http.MethodGet).Name("GetInstanceLogs")

	// Templates ..
	s.Router().HandleFunc("/api/action-templates/", s.handler.actionTemplateFolders).Methods(http.MethodGet).Name("ListActionTemplateFolders")
	s.Router().HandleFunc("/api/action-templates/{folder}/", s.handler.actionTemplates).Methods(http.MethodGet).Name("ListActionTemplates")
	s.Router().HandleFunc("/api/action-templates/{folder}/{template}", s.handler.actionTemplate).Methods(http.MethodGet).Name("GetActionTemplate")

	s.Router().HandleFunc("/api/workflow-templates/", s.handler.workflowTemplateFolders).Methods(http.MethodGet).Name("ListWorkflowTemplateFolders")
	s.Router().HandleFunc("/api/workflow-templates/{folder}/", s.handler.workflowTemplates).Methods(http.MethodGet).Name("ListWorkflowTemplates")
	s.Router().HandleFunc("/api/workflow-templates/{folder}/{template}", s.handler.workflowTemplate).Methods(http.MethodGet).Name("GetWorkflowTemplate")

	// jq Playground ...
	s.Router().HandleFunc("/api/jq-playground", s.handler.jqPlayground).Methods(http.MethodPost).Name("JQPlayground")

}

// Start starts the API server
func (s *Server) Start() error {

	log.Infof("Starting server - binding to %s", s.cfg.Server.Bind)

	if _, err := os.Stat(direktiv.TLSCert); !os.IsNotExist(err) {
		log.Infof("tls enabled")
		return s.srv.ListenAndServeTLS(direktiv.TLSCert, direktiv.TLSKey)
	}

	return s.srv.ListenAndServe()
}

func init() {
	resolver.Register(&direktiv.KubeResolverBuilder{})
}
