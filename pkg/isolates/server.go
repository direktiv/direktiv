package isolates

import (
	"context"
	"fmt"
	"net"
	"os"

	log "github.com/sirupsen/logrus"
	igrpc "github.com/vorteil/direktiv/pkg/isolates/grpc"
	"google.golang.org/grpc"
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

// func (is *isolateServer) UpdateIsolate(ctx context.Context,
// 	in *igrpc.CreateIsolateRequest) (*emptypb.Empty, error) {
//
// 	var empty = &emptypb.Empty{}
//
// 	// store db first
// 	// is.StoreIsolate(ctx, in)
//
// 	err := updateServiceKube(in.GetName(), in.GetNamespace(), in.GetWorkflow(),
// 		in.GetConfig(), in.GetExternal())
// 	if err != nil {
// 		log.Errorf("can not create knative service: %v", err)
// 		return empty, err
// 	}
//
// 	return empty, nil
//
// }

// StoreIsolate saves or updates isolates which means creating knative services
// baes on the provided configuration
func (is *isolateServer) CreateIsolate(ctx context.Context,
	in *igrpc.CreateIsolateRequest) (*emptypb.Empty, error) {

	log.Infof("storing isolate %s", in.GetInfo().GetName())

	if in.GetInfo() == nil || in.GetConfig() == nil {
		return &empty, fmt.Errorf("info and config can not be nil")
	}

	// create ksvc service
	err := createKnativeIsolate(in.GetInfo(), in.GetConfig(), in.GetExternal())
	if err != nil {
		log.Errorf("can not create knative service: %v", err)
		return &empty, err
	}

	return &empty, nil

}
