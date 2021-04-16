package api

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gorilla/mux"
	"github.com/vorteil/direktiv/pkg/ingress"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Server ..
type Server struct {
	cfg *Config
	// handlers map[*mux.Router][]*Handler
	direktiv ingress.DirektivIngressClient
	json     jsonpb.Marshaler
	handlers *Handlers
	routes   map[string]map[string]*Handler
	router   *mux.Router
	srv      *http.Server
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
	}

	s.handlers = &Handlers{
		s: s,
	}

	err := s.initDirektiv()
	if err != nil {
		return nil, err
	}

	s.prepareRoutes()

	return s, nil
}

func (s *Server) Handlers() *Handlers {
	return s.handlers
}

func (s *Server) initDirektiv() error {

	var opts []grpc.DialOption
	if s.cfg.Ingress.TLS.Enabled {
		tc, err := tlsConfig(s.cfg.Ingress.TLS.CertsDir, "client", s.cfg.Ingress.TLS.Secure)
		if err != nil {
			return err
		}

		if len(tc.Certificates) > 0 {
			opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tc)))
		} else {
			opts = append(opts, grpc.WithInsecure())
		}
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	conn, err := grpc.Dial(s.cfg.Ingress.Endpoint, opts...)
	if err != nil {
		return err
	}

	s.direktiv = ingress.NewDirektivIngressClient(conn)
	return nil
}

func (s *Server) prepareRoutes() {

	// init routes map
	s.routes = make(map[string]map[string]*Handler)
	s.routes[http.MethodGet] = make(map[string]*Handler)
	s.routes[http.MethodPost] = make(map[string]*Handler)
	s.routes[http.MethodPut] = make(map[string]*Handler)
	s.routes[http.MethodDelete] = make(map[string]*Handler)
	s.routes[http.MethodPatch] = make(map[string]*Handler)
	s.routes[http.MethodOptions] = make(map[string]*Handler)

	// Options ..
	s.routes[http.MethodOptions]["/{path:.*}"] = NewHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Webhook-Allowed-Origin", "eventgrid.azure.net")
		w.WriteHeader(http.StatusOK)
	})

	// Namespace ..
	s.routes[http.MethodGet]["/api/namespaces/"] = NewHandler(s.handlers.Namespaces)
	s.routes[http.MethodPost]["/api/namespaces/{namespace}"] = NewHandler(s.handlers.AddNamespace)
	s.routes[http.MethodDelete]["/api/namespaces/{namespace}"] = NewHandler(s.handlers.DeleteNamespace)

	// Event ..
	s.routes[http.MethodPost]["/api/namespaces/{namespace}/event"] = NewHandler(s.handlers.NamespaceEvent)

	// Secret ..
	s.routes[http.MethodGet]["/api/namespaces/{namespace}/secrets/"] = NewHandler(s.handlers.Secrets)
	s.routes[http.MethodPost]["/api/namespaces/{namespace}/secrets/"] = NewHandler(s.handlers.CreateSecret)
	s.routes[http.MethodDelete]["/api/namespaces/{namespace}/secrets/"] = NewHandler(s.handlers.DeleteSecret)

	// Registry ..
	s.routes[http.MethodGet]["/api/namespaces/{namespace}/registries/"] = NewHandler(s.handlers.Registries)
	s.routes[http.MethodPost]["/api/namespaces/{namespace}/registries/"] = NewHandler(s.handlers.CreateRegistry)
	s.routes[http.MethodDelete]["/api/namespaces/{namespace}/registries/"] = NewHandler(s.handlers.DeleteRegistry)

	// Workflow ..
	s.routes[http.MethodGet]["/api/namespaces/{namespace}/workflows/"] = NewHandler(s.handlers.Workflows)
	s.routes[http.MethodGet]["/api/namespaces/{namespace}/workflows/{workflowTarget}"] = NewHandler(s.handlers.GetWorkflow)
	s.routes[http.MethodPut]["/api/namespaces/{namespace}/workflows/{workflowUID}"] = NewHandler(s.handlers.UpdateWorkflow)
	s.routes[http.MethodPut]["/api/namespaces/{namespace}/workflows/{workflowUID}/toggle"] = NewHandler(s.handlers.ToggleWorkflow)
	s.routes[http.MethodPost]["/api/namespaces/{namespace}/workflows"] = NewHandler(s.handlers.CreateWorkflow)
	s.routes[http.MethodDelete]["/api/namespaces/{namespace}/workflows/{workflowUID}"] = NewHandler(s.handlers.DeleteWorkflow)
	s.routes[http.MethodGet]["/api/namespaces/{namespace}/workflows/{workflowUID}/download"] = NewHandler(s.handlers.DownloadWorkflow)
	s.routes[http.MethodPost]["/api/namespaces/{namespace}/workflows/{workflowID}/execute"] = NewHandler(s.handlers.ExecuteWorkflow)

	// Instance ..
	s.routes[http.MethodGet]["/api/instances/{namespace}"] = NewHandler(s.handlers.Instances)
	s.routes[http.MethodGet]["/api/instances/{namespace}/{workflowID}/{id}"] = NewHandler(s.handlers.GetInstance)
	s.routes[http.MethodDelete]["/api/instances/{namespace}/{workflowID}/{id}"] = NewHandler(s.handlers.CancelInstance)
	s.routes[http.MethodGet]["/api/instances/{namespace}/{workflowID}/{id}/logs"] = NewHandler(s.handlers.InstanceLogs)

}

func (s *Server) RegisterHandler(path string, h *Handler, methods ...string) {
	for _, method := range methods {
		s.routes[method][path] = h
	}
}

func (s *Server) Start() error {

	for method, paths := range s.routes {
		for path, h := range paths {
			s.router.HandleFunc(path, h.exec).Methods(method)
		}
	}

	if s.cfg.Server.TLS.Enabled {
		fmt.Println("Starting TLS server...")
		return s.srv.ListenAndServeTLS(filepath.Join(s.cfg.Server.TLS.CertsDir, "cert.pem"), filepath.Join(s.cfg.Server.TLS.CertsDir, "key.pem"))
	}

	fmt.Printf("Starting server - binding to %s!\n", s.cfg.Server.Bind)
	return s.srv.ListenAndServe()
}
