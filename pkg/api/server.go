package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gorilla/mux"
	prometheus "github.com/prometheus/client_golang/api"
	"github.com/vorteil/direktiv/pkg/dlog"
	igrpc "github.com/vorteil/direktiv/pkg/functions/grpc"
	"github.com/vorteil/direktiv/pkg/ingress"
	"github.com/vorteil/direktiv/pkg/util"
	"go.uber.org/zap"
)

const (
	PROMETHEUS_ADDR_ENV = "PROMETHEUS_ADDR"
)

// Server ..
type Server struct {
	cfg       *Config
	direktiv  ingress.DirektivIngressClient
	functions igrpc.FunctionsServiceClient
	json      jsonpb.Marshaler
	handler   *Handler
	router    *mux.Router
	srv       *http.Server

	reqMapMutex sync.Mutex
	reqMap      map[*http.Request]*RequestStatus

	wfTemplateDirsPaths map[string]string
	wfTemplateDirs      []string

	actionTemplateDirsPaths map[string]string
	actionTemplateDirs      []string

	blocklist         []string
	prometheusEnabled bool
	prometheus        prometheus.Client
}

var logger *zap.SugaredLogger

// NewServer returns new API server
func NewServer() (*Server, error) {

	var err error

	logger, err = dlog.ApplicationLogger("api")
	if err != nil {
		return nil, err
	}

	cfg, err := Configure()
	if err != nil {
		return nil, err
	}

	r := mux.NewRouter()
	var bl []string

	logger.Infof("check for a blocklist")

	if cfg.hasBlockList() {
		logger.Infof("contains a blocklist")
		// fetch blocklist
		data, err := ioutil.ReadFile(cfg.BlockList)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(data, &bl)
		if err != nil {
			return nil, err
		}

		logger.Infof("blocklist %s", data)
	}

	s := &Server{
		cfg:    cfg,
		router: r,
		srv: &http.Server{
			Handler: r,
			Addr:    apiBind,
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

	err = s.initFunctions()
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

// FunctionsClient returns client to backend
func (s *Server) FunctionsClient() igrpc.FunctionsServiceClient {
	return s.functions
}

// Router returns mux router
func (s *Server) Router() *mux.Router {
	return s.router
}

func (s *Server) initDirektiv() error {

	conn, err := util.GetEndpointTLS(util.TLSIngressComponent)
	if err != nil {
		logger.Errorf("can not connect to direktiv ingress: %v", err)
		return err
	}

	logger.Infof("connecting to %s", util.IngressEndpoint())

	s.direktiv = ingress.NewDirektivIngressClient(conn)

	return nil
}

func (s *Server) initFunctions() error {

	conn, err := util.GetEndpointTLS(util.TLSFunctionsComponent)
	if err != nil {
		logger.Errorf("can not connect to direktiv functions: %v", err)
		return err
	}

	logger.Infof("connecting to %s", util.FunctionsEndpoint())

	s.functions = igrpc.NewFunctionsServiceClient(conn)

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

	//Testing
	s.Router().HandleFunc("/api/watch/instance/{namespace}/{workflowTarget}/{id}", s.handler.watchInstanceLogs).Methods(http.MethodGet).Name(RN_WatchInstanceServices)

	// Watch
	s.Router().HandleFunc("/api/watch/functions/", s.handler.watchFunctions).Methods(http.MethodGet).Name(RN_WatchServices)
	s.Router().HandleFunc("/api/watch/functions/{serviceName}", s.handler.watchFunctions).Methods(http.MethodGet).Name(RN_WatchServices)
	s.Router().HandleFunc("/api/watch/functions/{serviceName}/revisions/", s.handler.watchRevisions).Methods(http.MethodGet).Name(RN_WatchRevisions)
	s.Router().HandleFunc("/api/watch/functions/{serviceName}/revisions/{revisionName}", s.handler.watchRevisions).Methods(http.MethodGet).Name(RN_WatchRevisions)

	s.Router().HandleFunc("/api/watch/functions/{serviceName}/revisions/{revisionName}/pods/", s.handler.watchPods).Methods(http.MethodGet).Name(RN_WatchPods)
	s.Router().HandleFunc("/api/watch/functions/{serviceName}/revisions/{revisionName}/pods/{podName}/logs/", s.handler.watchLogs).Methods(http.MethodGet).Name(RN_WatchLogs)

	s.Router().HandleFunc("/api/watch/namespaces/{namespace}/functions/", s.handler.watchFunctions).Methods(http.MethodGet).Name(RN_WatchNamespaceServices)
	s.Router().HandleFunc("/api/watch/namespaces/{namespace}/functions/{serviceName}", s.handler.watchFunctions).Methods(http.MethodGet).Name(RN_WatchNamespaceServices)
	s.Router().HandleFunc("/api/watch/namespaces/{namespace}/functions/{serviceName}/revisions/", s.handler.watchRevisions).Methods(http.MethodGet).Name(RN_WatchNamespaceRevisions)
	s.Router().HandleFunc("/api/watch/namespaces/{namespace}/functions/{serviceName}/revisions/{revisionName}", s.handler.watchRevisions).Methods(http.MethodGet).Name(RN_WatchNamespaceRevisions)

	s.Router().HandleFunc("/api/watch/namespaces/{namespace}/functions/{serviceName}/revisions/{revisionName}/pods/", s.handler.watchPods).Methods(http.MethodGet).Name(RN_WatchPods)
	s.Router().HandleFunc("/api/watch/namespaces/{namespace}/functions/{serviceName}/revisions/{revisionName}/pods/{podName}/logs/", s.handler.watchLogs).Methods(http.MethodGet).Name(RN_WatchLogs)

	s.Router().HandleFunc("/api/watch/namespaces/{namespace}/workflows/{workflowTarget}/functions/", s.handler.watchFunctions).Methods(http.MethodGet).Name(RN_WatchNamespaceServices)

	// Functions ..
	s.Router().HandleFunc("/api/functions/", s.handler.listServices).Methods(http.MethodPost).Name(RN_ListServices)
	s.Router().HandleFunc("/api/functions/pods/", s.handler.listPods).Methods(http.MethodPost).Name(RN_ListPods)
	s.Router().HandleFunc("/api/functions/", s.handler.deleteServices).Methods(http.MethodDelete).Name(RN_DeleteServices)
	s.Router().HandleFunc("/api/functions/new", s.handler.createService).Methods(http.MethodPost).Name(RN_CreateService)
	s.Router().HandleFunc("/api/functions/{serviceName}", s.handler.getService).Methods(http.MethodGet).Name(RN_GetService)
	s.Router().HandleFunc("/api/functions/{serviceName}", s.handler.updateService).Methods(http.MethodPost).Name(RN_UpdateService)
	s.Router().HandleFunc("/api/functions/{serviceName}", s.handler.updateServiceTraffic).Methods(http.MethodPatch).Name(RN_UpdateServiceTraffic)
	s.Router().HandleFunc("/api/functions/{serviceName}", s.handler.deleteService).Methods(http.MethodDelete).Name(RN_DeleteService)
	s.Router().HandleFunc("/api/functionrevisions/{revision}", s.handler.deleteRevision).Methods(http.MethodDelete).Name(RN_DeleteRevision)

	s.Router().HandleFunc("/api/namespaces/{namespace}/metrics/workflows-invoked", s.handler.getNamespaceMetrics_WorkflowsInvoked).Methods(http.MethodGet).Name(RN_namespaceWorkflowsInvoked)
	s.Router().HandleFunc("/api/namespaces/{namespace}/metrics/workflows-successful", s.handler.getNamespaceMetrics_WorkflowsSuccessful).Methods(http.MethodGet).Name(RN_namespaceWorkflowsSuccessful)
	s.Router().HandleFunc("/api/namespaces/{namespace}/metrics/workflows-failed", s.handler.getNamespaceMetrics_WorkflowsFailed).Methods(http.MethodGet).Name(RN_namespaceWorkflowsFailed)
	s.Router().HandleFunc("/api/namespaces/{namespace}/metrics/workflows-milliseconds", s.handler.getNamespaceMetrics_WorkflowsMilliseconds).Methods(http.MethodGet).Name(RN_namespaceWorkflowsMS)

	s.Router().HandleFunc("/api/namespaces/{namespace}/workflows/{workflow}/metrics/invoked", s.handler.getWorkflowMetrics_Invoked).Methods(http.MethodGet).Name(RN_metricsWorkflowInvoked)
	s.Router().HandleFunc("/api/namespaces/{namespace}/workflows/{workflow}/metrics/successful", s.handler.getWorkflowMetrics_Successful).Methods(http.MethodGet).Name(RN_metricsWorkflowSuccessful)
	s.Router().HandleFunc("/api/namespaces/{namespace}/workflows/{workflow}/metrics/failed", s.handler.getWorkflowMetrics_Failed).Methods(http.MethodGet).Name(RN_metricsWorkflowFailed)
	s.Router().HandleFunc("/api/namespaces/{namespace}/workflows/{workflow}/metrics/milliseconds", s.handler.getWorkflowMetrics_Milliseconds).Methods(http.MethodGet).Name(RN_metricsWorkflowMS)
	s.Router().HandleFunc("/api/namespaces/{namespace}/workflows/{workflow}/metrics/state-milliseconds", s.handler.getWorkflowMetrics_StateMilliseconds).Methods(http.MethodGet).Name(RN_metricsStateMS)

	s.Router().HandleFunc("/api/namespaces/{namespace}/functions/", s.handler.listServices).Methods(http.MethodPost).Name(RN_ListServices)
	s.Router().HandleFunc("/api/namespaces/{namespace}/functions/pods/", s.handler.listPods).Methods(http.MethodPost).Name(RN_ListPods)
	s.Router().HandleFunc("/api/namespaces/{namespace}/functions/", s.handler.deleteServices).Methods(http.MethodDelete).Name(RN_DeleteServices)
	s.Router().HandleFunc("/api/namespaces/{namespace}/functions/new", s.handler.createService).Methods(http.MethodPost).Name(RN_CreateService)
	s.Router().HandleFunc("/api/namespaces/{namespace}/functions/{serviceName}", s.handler.getService).Methods(http.MethodGet).Name(RN_GetService)
	s.Router().HandleFunc("/api/namespaces/{namespace}/functions/{serviceName}", s.handler.updateService).Methods(http.MethodPost).Name(RN_UpdateService)
	s.Router().HandleFunc("/api/namespaces/{namespace}/functions/{serviceName}", s.handler.updateServiceTraffic).Methods(http.MethodPatch).Name(RN_UpdateServiceTraffic)
	s.Router().HandleFunc("/api/namespaces/{namespace}/functions/{serviceName}", s.handler.deleteService).Methods(http.MethodDelete).Name(RN_DeleteService)
	s.Router().HandleFunc("/api/namespaces/{namespace}/functionrevisions/{revision}", s.handler.deleteRevision).Methods(http.MethodDelete).Name(RN_DeleteRevision)

	s.Router().HandleFunc("/api/namespaces/{namespace}/workflows/{workflowTarget}/functions", s.handler.getWorkflowFunctions).Methods(http.MethodGet).Name(RN_GetWorkflowFunctions)

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

// Start starts the API server
func (s *Server) Start() error {

	logger.Infof("Starting server - binding to %s", apiBind)

	var err error

	if os.Getenv(PROMETHEUS_ADDR_ENV) != "" {
		s.prometheus, err = prometheus.NewClient(prometheus.Config{
			Address: os.Getenv(PROMETHEUS_ADDR_ENV),
		})
		if err != nil {
			return err
		}

		s.prometheusEnabled = true
	}

	k, c, _ := util.CertsForComponent(util.TLSHttpComponent)
	if len(k) > 0 {
		logger.Infof("api tls enabled")
		return s.srv.ListenAndServeTLS(c, k)
	}

	return s.srv.ListenAndServe()
}
