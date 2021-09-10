package flow

import (
	"context"
	"net"

	"github.com/vorteil/direktiv/pkg/flow/grpc"
	libgrpc "google.golang.org/grpc"
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

	flow.listener, err = net.Listen("tcp", srv.conf.Bind)
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
