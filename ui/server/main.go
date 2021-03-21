package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

var (
	bind             string
	grpcAddr         string
	direktivCertsDir string
	tlsCertsDir      string
	insecure         bool
	webDir           string
)

func init() {
	flag.StringVar(&bind, "bind", ":8080", "api server bind address")
	flag.StringVar(&grpcAddr, "direktiv-endpoint", "127.0.0.1:6666", "direktiv ingress grpc server endpoint")
	flag.StringVar(&direktivCertsDir, "direktiv-certs-dir", "", "directory containing direktiv grpc TLS certificates")
	flag.StringVar(&tlsCertsDir, "certs-dir", "", "directory containing TLS certificates for api server")
	flag.BoolVar(&insecure, "insecure", true, "skip certificate verification")
	flag.StringVar(&webDir, "web-dir", "web", "directory to serve web files from")
	flag.Parse()
}

func main() {

	if len(os.Args) > 1 && os.Args[1] == "help" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	s := http.Server{}
	s.Addr = bind

	r := mux.NewRouter()
	s.Handler = r

	gc := new(grpcClient)
	gc.addr = grpcAddr

	err := gc.init()
	if err != nil {
		panic(err)
	}

	// Get ...
	r.HandleFunc("/api/namespaces", gc.namespacesHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/namespaces", gc.createNamespaceHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/namespaces/{namespace}/workflows", gc.workflowsHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/namespaces/{namespace}/workflows", gc.createWorkflowHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/namespaces/{namespace}/workflows/{workflow}", gc.getWorkflowHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/instances/{namespace}", gc.instancesHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/instances/{namespace}/{workflowID}/{id}", gc.instanceHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/instances/{namespace}/{workflowID}/{id}/logs", gc.instanceLogsHandler).Methods(http.MethodGet)

	// Post ...
	r.HandleFunc("/api/namespaces/{namespace}/workflows/{workflow}/execute", gc.executeWorkflowHandler).Methods(http.MethodPost)

	// Web Handler
	// r.Handle("/build/web/", http.StripPrefix("/build/web/", http.FileServer(http.Dir("build/web"))))
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(webDir)))

	fmt.Printf(`Starting API Server 
  -bind='%s'
  -web-dir='%s'
  -direktiv-endpoint='%s'
  -direktiv-certs-dir='%s'
  -certs-dir='%s'
  -insecure='%v'
`, bind, webDir, grpcAddr, direktivCertsDir, tlsCertsDir, insecure)

	err = s.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
