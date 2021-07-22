package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/pkg/direktiv"
	"github.com/vorteil/direktiv/pkg/ingress"
)

const blocklist = "blocklist"

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

	blocklist []string
}

// NewServer returns new API server
func NewServer(cfg *Config) (*Server, error) {

	r := mux.NewRouter()

	// fetch blocklist
	var bl []string
	data, err := ioutil.ReadFile(blocklist)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &bl)
	if err != nil {
		return nil, err
	}

	log.Infof("blocklist %s", data)

	s := &Server{
		cfg:    cfg,
		router: r,
		srv: &http.Server{
			Handler: r,
			Addr:    cfg.Server.Bind,
		},
		blocklist:   bl,
		reqMapMutex: sync.Mutex{},
		reqMap:      make(map[*http.Request]*RequestStatus),
		json: jsonpb.Marshaler{
			EmitDefaults: true,
		},
	}

	s.handler = &Handler{
		s: s,
	}

	err = s.initDirektiv()
	if err != nil {
		return nil, err
	}

	err = s.initTemplateFolders()
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

	conn, err := direktiv.GetEndpointTLS(s.cfg.Ingress.Endpoint, true)
	if err != nil {
		log.Errorf("can not connect to direktiv ingress: %v", err)
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
	}).Methods(http.MethodOptions).Name(RN_Preflight)

	// Health check
	s.Router().HandleFunc("/api/healthz", func(w http.ResponseWriter, r *http.Request) {
		// responds 200 OK
	}).Methods(http.MethodGet).Name(RN_HealthCheck)

	// Namespace ..
	s.Router().HandleFunc("/api/namespaces/", s.handler.namespaces).Methods(http.MethodGet).Name(RN_ListNamespaces)
	s.Router().HandleFunc("/api/namespaces/{namespace}", s.handler.addNamespace).Methods(http.MethodPost).Name(RN_AddNamespace)
	s.Router().HandleFunc("/api/namespaces/{namespace}", s.handler.deleteNamespace).Methods(http.MethodDelete).Name(RN_DeleteNamespace)

	// Logs ..
	s.Router().HandleFunc("/api/namespaces/{namespace}/logs", s.handler.namespaceLogs).Methods(http.MethodGet).Name(RN_GetNamespaceLogs)

	// Event ..
	s.Router().HandleFunc("/api/namespaces/{namespace}/event", s.handler.namespaceEvent).Methods(http.MethodPost).Name(RN_NamespaceEvent)

	// Secret ..
	s.Router().HandleFunc("/api/namespaces/{namespace}/secrets/", s.handler.getSecretsOrRegistries).Methods(http.MethodGet).Name(RN_ListSecrets)
	s.Router().HandleFunc("/api/namespaces/{namespace}/secrets/", s.handler.createSecretOrRegistry).Methods(http.MethodPost).Name(RN_CreateSecret)
	s.Router().HandleFunc("/api/namespaces/{namespace}/secrets/", s.handler.deleteSecretOrRegistry).Methods(http.MethodDelete).Name(RN_DeleteSecret)

	// Registry ..
	s.Router().HandleFunc("/api/namespaces/{namespace}/registries/", s.handler.getSecretsOrRegistries).Methods(http.MethodGet).Name(RN_ListRegistries)
	s.Router().HandleFunc("/api/namespaces/{namespace}/registries/", s.handler.createSecretOrRegistry).Methods(http.MethodPost).Name(RN_CreateRegistry)
	s.Router().HandleFunc("/api/namespaces/{namespace}/registries/", s.handler.deleteSecretOrRegistry).Methods(http.MethodDelete).Name(RN_DeleteRegistry)

	// Metrics ..
	s.Router().HandleFunc("/api/namespaces/{namespace}/workflows/{workflow}/metrics", s.handler.workflowMetrics).Methods(http.MethodGet).Name(RN_GetWorkflowMetrics)

	// Workflow ..
	s.Router().HandleFunc("/api/namespaces/{namespace}/workflows/", s.handler.workflows).Methods(http.MethodGet).Name(RN_ListWorkflows)
	s.Router().HandleFunc("/api/namespaces/{namespace}/workflows/{workflowTarget}", s.handler.getWorkflow).Methods(http.MethodGet).Name(RN_GetWorkflow)
	s.Router().HandleFunc("/api/namespaces/{namespace}/workflows/{workflowTarget}", s.handler.updateWorkflow).Methods(http.MethodPut).Name(RN_UpdateWorkflow)
	s.Router().HandleFunc("/api/namespaces/{namespace}/workflows/{workflowTarget}/toggle", s.handler.toggleWorkflow).Methods(http.MethodPut).Name(RN_ToggleWorkflow)
	s.Router().HandleFunc("/api/namespaces/{namespace}/workflows", s.handler.createWorkflow).Methods(http.MethodPost).Name(RN_CreateWorkflow)
	s.Router().HandleFunc("/api/namespaces/{namespace}/workflows/{workflowTarget}", s.handler.deleteWorkflow).Methods(http.MethodDelete).Name(RN_DeleteWorkflow)
	s.Router().HandleFunc("/api/namespaces/{namespace}/workflows/{workflowTarget}/download", s.handler.downloadWorkflow).Methods(http.MethodGet).Name(RN_DownloadWorkflow)
	s.Router().HandleFunc("/api/namespaces/{namespace}/workflows/{workflowTarget}/execute", s.handler.executeWorkflow).Methods(http.MethodPost, http.MethodGet).Name(RN_ExecuteWorkflow)
	s.Router().HandleFunc("/api/namespaces/{namespace}/workflows/{workflowTarget}/instances/", s.handler.workflowInstances).Methods(http.MethodGet).Name(RN_ListWorkflowInstances)

	// Instance ..
	s.Router().HandleFunc("/api/instances/{namespace}", s.handler.instances).Methods(http.MethodGet).Name(RN_ListInstances)
	s.Router().HandleFunc("/api/instances/{namespace}/{workflowTarget}/{id}", s.handler.getInstance).Methods(http.MethodGet).Name(RN_GetInstance)
	s.Router().HandleFunc("/api/instances/{namespace}/{workflowTarget}/{id}", s.handler.cancelInstance).Methods(http.MethodDelete).Name(RN_CancelInstance)
	s.Router().HandleFunc("/api/instances/{namespace}/{workflowTarget}/{id}/logs", s.handler.instanceLogs).Methods(http.MethodGet).Name(RN_GetInstanceLogs)

	// Templates ..
	s.Router().HandleFunc("/api/action-templates/", s.handler.templateFolders).Methods(http.MethodGet).Name(RN_ListActionTemplateFolders)
	s.Router().HandleFunc("/api/action-templates/{folder}/", s.handler.actionTemplates).Methods(http.MethodGet).Name(RN_ListActionTemplates)
	s.Router().HandleFunc("/api/action-templates/{folder}/{template}", s.handler.getTemplate).Methods(http.MethodGet).Name(RN_GetActionTemplate)

	s.Router().HandleFunc("/api/workflow-templates/", s.handler.templateFolders).Methods(http.MethodGet).Name(RN_ListWorkflowTemplateFolders)
	s.Router().HandleFunc("/api/workflow-templates/{folder}/", s.handler.workflowTemplates).Methods(http.MethodGet).Name(RN_ListWorkflowTemplates)
	s.Router().HandleFunc("/api/workflow-templates/{folder}/{template}", s.handler.getTemplate).Methods(http.MethodGet).Name(RN_GetWorkflowTemplate)

	// Varaibles
	s.Router().HandleFunc("/api/namespaces/{namespace}/workflows/{workflowTarget}/variables/", s.handler.workflowVariables).Methods(http.MethodGet).Name(RN_ListWorkflowVariables)
	s.Router().HandleFunc("/api/namespaces/{namespace}/workflows/{workflowTarget}/variables/{variable}", s.handler.setWorkflowVariable).Methods(http.MethodPost).Name(RN_SetWorkflowVariable)
	s.Router().HandleFunc("/api/namespaces/{namespace}/workflows/{workflowTarget}/variables/{variable}", s.handler.getWorkflowVariable).Methods(http.MethodGet).Name(RN_GetWorkflowVariable)
	s.Router().HandleFunc("/api/namespaces/{namespace}/variables/", s.handler.namespaceVariables).Methods(http.MethodGet).Name(RN_ListNamespaceVariables)
	s.Router().HandleFunc("/api/namespaces/{namespace}/variables/{variable}", s.handler.setNamespaceVariable).Methods(http.MethodPost).Name(RN_SetNamespaceVariable)
	s.Router().HandleFunc("/api/namespaces/{namespace}/variables/{variable}", s.handler.getNamespaceVariable).Methods(http.MethodGet).Name(RN_GetNamespaceVariable)

	// jq Playground ...
	s.Router().HandleFunc("/api/jq-playground", s.handler.jqPlayground).Methods(http.MethodPost).Name(RN_JQPlayground)

}

const tlsDir = "/etc/certs/servedirektiv"

func tlsEnabled() bool {
	_, err := os.Stat(tlsDir)
	return !os.IsNotExist(err)
}

// Start starts the API server
func (s *Server) Start() error {

	log.Infof("Starting server - binding to %s", s.cfg.Server.Bind)

	if tlsEnabled() {
		log.Infof("api tls enabled")
		return s.srv.ListenAndServeTLS(filepath.Join(tlsDir, "tls.crt"), filepath.Join(tlsDir, "tls.key"))
	}

	return s.srv.ListenAndServe()
}
