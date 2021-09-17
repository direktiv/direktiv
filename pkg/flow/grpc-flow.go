package flow

import (
	"context"
	"net"

	"github.com/vorteil/direktiv/pkg/flow/grpc"
	libgrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type flow struct {
	*server
	listener net.Listener
	srv      *libgrpc.Server
	grpc.UnimplementedFlowServer
}

func initFlowServer(ctx context.Context, srv *server) (*flow, error) {

	var err error

	flow := &flow{server: srv}

	flow.listener, err = net.Listen("tcp", ":6666")
	if err != nil {
		return nil, err
	}

	flow.srv = libgrpc.NewServer(libgrpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *libgrpc.UnaryServerInfo, handler libgrpc.UnaryHandler) (resp interface{}, err error) {
		resp, err = handler(ctx, req)
		if err != nil {
			return nil, translateError(err)
		}
		return resp, nil
	}))

	grpc.RegisterFlowServer(flow.srv, flow)
	reflection.Register(flow.srv)

	go func() {
		<-ctx.Done()
		flow.srv.Stop()
	}()

	return flow, nil

}

func (flow *flow) Run() error {

	err := flow.srv.Serve(flow.listener)
	if err != nil {
		return err
	}

	return nil

}
