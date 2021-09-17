package util

import (
	"net"

	"google.golang.org/grpc"
)

const maxSize = 33554432

// GetEndpointTLS creates a grpc client
func GetEndpointTLS(service string) (*grpc.ClientConn, error) {

	var additionalCallOptions []grpc.CallOption
	additionalCallOptions = append(additionalCallOptions, grpc.MaxCallSendMsgSize(maxSize))
	additionalCallOptions = append(additionalCallOptions, grpc.MaxCallRecvMsgSize(maxSize))

	var options []grpc.DialOption

	options = append(options,
		grpc.WithDefaultCallOptions(additionalCallOptions...))

	return grpc.Dial(service, options...)

}

// GrpcStart starts a grpc server
func GrpcStart(server **grpc.Server, name, bind string, register func(srv *grpc.Server)) error {

	listener, err := net.Listen("tcp", bind)
	if err != nil {
		return err
	}

	var additionalServerOptions []grpc.ServerOption
	additionalServerOptions = append(additionalServerOptions, grpc.MaxSendMsgSize(maxSize))
	additionalServerOptions = append(additionalServerOptions, grpc.MaxRecvMsgSize(maxSize))

	(*server) = grpc.NewServer(additionalServerOptions...)

	register(*server)

	go (*server).Serve(listener)

	return nil

}
