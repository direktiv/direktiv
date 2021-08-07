package isolates

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	igrpc "github.com/vorteil/direktiv/pkg/isolates/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	grpcServer    *grpc.Server
	empty         emptypb.Empty
	isolateConfig config
)

const (
	confFile = "/etc/direktiv/config.yaml"
	port     = 5555
)

type isolateServer struct {
	igrpc.UnimplementedIsolatesServiceServer
}

// StartServer starts isolate grpc server
func StartServer(echan chan error) {

	errChan := make(chan error)
	go runPodRequestLimiter(errChan)

	e := <-errChan
	if e != nil {
		echan <- e
		return
	}

	err := initKubernetesLock()
	if err != nil {
		echan <- err
	}

	cr := newConfigReader()
	go cr.readConfig(confFile, &isolateConfig)

	if len(os.Getenv(envFlow)) == 0 {
		log.Errorf("grpc response is not configured (DIREKTIV_FLOW_ENDPOINT)")
		echan <- fmt.Errorf("grpc response is not configured (DIREKTIV_FLOW_ENDPOINT)")
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Errorf("failed to listen: %v", err)
		echan <- err
	}

	var opts []grpc.ServerOption
	tlsPath := "/etc/certs/direktiv/"
	tlsCert := filepath.Join(tlsPath, "tls.crt")
	tlsKey := filepath.Join(tlsPath, "tls.key")

	if _, err = os.Stat(tlsKey); err == nil {
		log.Infof("enabling tls for %s", "isolates")
		creds, err := credentials.NewServerTLSFromFile(tlsCert, tlsKey)
		if err != nil {
			log.Errorf("failed to configure tls opts: %v", err)
			echan <- fmt.Errorf("could not load TLS keys: %s", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	grpcServer = grpc.NewServer(opts...)
	igrpc.RegisterIsolatesServiceServer(grpcServer, &isolateServer{})

	if err := grpcServer.Serve(lis); err != nil {
		echan <- err
	}

}

// StopServer is stopping server gracefully
func StopServer() {
	if grpcServer != nil {
		grpcServer.GracefulStop()
	}
}
