package isolates

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	igrpc "github.com/vorteil/direktiv/pkg/isolates/grpc"
	"github.com/vorteil/direktiv/pkg/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	grpcServer    *grpc.Server
	empty         emptypb.Empty
	isolateConfig config
)

const (
	confFile = "/etc/direktiv/isolate-config.yaml"
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
		return
	}

	cr := newConfigReader()
	go cr.readConfig(confFile, &isolateConfig)

	if len(util.FlowEndpoint()) == 0 {
		log.Errorf("grpc response to flow is not configured")
		echan <- fmt.Errorf("grpc response to flow is not configured")
	}

	err = util.GrpcStart(&grpcServer, util.TLSIsolatesComponent,
		fmt.Sprintf(":%d", port), func(srv *grpc.Server) {
			igrpc.RegisterIsolatesServiceServer(srv, &isolateServer{})
			reflection.Register(srv)
		})

	if err != nil {
		echan <- err
	}

	select {}

}

// StopServer is stopping server gracefully
func StopServer() {
	if grpcServer != nil {
		grpcServer.GracefulStop()
	}
}
