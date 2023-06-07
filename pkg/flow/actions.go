package flow

import (
	"context"
	"net"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/grpc"
	igrpc "github.com/direktiv/direktiv/pkg/functions/grpc"
	"github.com/direktiv/direktiv/pkg/util"
	libgrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type actions struct {
	*server
	listener net.Listener
	srv      *libgrpc.Server
	grpc.UnimplementedActionsServer

	conn   *libgrpc.ClientConn
	client igrpc.FunctionsClient
}

func initActionsServer(ctx context.Context, srv *server) (*actions, error) {
	var err error

	actions := &actions{server: srv}

	actions.conn, err = util.GetEndpointTLS(srv.conf.FunctionsService + ":5555")
	if err != nil {
		return nil, err
	}

	actions.client = igrpc.NewFunctionsClient(actions.conn)

	actions.listener, err = net.Listen("tcp", ":4444")
	if err != nil {
		return nil, err
	}

	opts := util.GrpcServerOptions(unaryInterceptor, streamInterceptor)

	actions.srv = libgrpc.NewServer(opts...)

	grpc.RegisterActionsServer(actions.srv, actions)
	reflection.Register(actions.srv)

	go func() {
		<-ctx.Done()
		actions.srv.Stop()
	}()

	return actions, nil
}

func (actions *actions) Run() error {
	err := actions.srv.Serve(actions.listener)
	if err != nil {
		return err
	}

	return nil
}

func (actions *actions) CancelWorkflowInstance(svn, actionID string) error {
	actions.sugar.Debugf("Handling gRPC request: %s", this())

	req := &igrpc.FunctionsCancelWorkflowRequest{
		ServiceName: &svn,
		ActionID:    &actionID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	_, err := actions.client.CancelWorfklow(ctx, req)

	return err
}
