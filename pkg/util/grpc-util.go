package util

import (
	"net"
	"strings"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
)

const maxSize = 33554432

// GetEndpointTLS creates a grpc client
func GetEndpointTLS(service string) (*grpc.ClientConn, error) {

	var additionalCallOptions []grpc.CallOption
	additionalCallOptions = append(additionalCallOptions, grpc.MaxCallSendMsgSize(maxSize))
	additionalCallOptions = append(additionalCallOptions, grpc.MaxCallRecvMsgSize(maxSize),
		grpc_retry.WithMax(10),
		grpc_retry.WithPerRetryTimeout(1*time.Second))

	var options []grpc.DialOption
	options = append(options, grpc.WithInsecure(), grpc.WithBlock())

	options = append(options,
		grpc.WithDefaultCallOptions(additionalCallOptions...),
	)

	options = append(options, globalGRPCDialOptions...)

	return grpc.Dial(service, options...)

}

// GrpcStart starts a grpc server
func GrpcStart(server **grpc.Server, name, bind string, register func(srv *grpc.Server)) error {

	listener, err := net.Listen("tcp", bind)
	if err != nil {
		return err
	}

	additionalServerOptions := GrpcServerOptions(nil, nil)

	(*server) = grpc.NewServer(additionalServerOptions...)

	register(*server)

	go (*server).Serve(listener)

	return nil

}

func GrpcServerOptions(unaryInterceptor grpc.UnaryServerInterceptor, streamInterceptor grpc.StreamServerInterceptor) []grpc.ServerOption {

	var additionalServerOptions []grpc.ServerOption
	additionalServerOptions = append(additionalServerOptions, grpc.MaxSendMsgSize(maxSize))
	additionalServerOptions = append(additionalServerOptions, grpc.MaxRecvMsgSize(maxSize))
	additionalServerOptions = append(additionalServerOptions, globalGRPCServerOptions...)

	// unary interceptors
	var unaryInterceptors []grpc.UnaryServerInterceptor
	if telemetryUnaryServerInterceptor != nil {
		unaryInterceptors = append(unaryInterceptors, telemetryUnaryServerInterceptor)
	}
	if unaryInterceptor != nil {
		unaryInterceptors = append(unaryInterceptors, unaryInterceptor)
	}

	l := len(unaryInterceptors)
	if l == 1 {
		additionalServerOptions = append(additionalServerOptions, grpc.UnaryInterceptor(unaryInterceptors[0]))
	} else if l > 1 {
		additionalServerOptions = append(additionalServerOptions, grpc.ChainUnaryInterceptor(unaryInterceptors...))
	}

	// stream interceptors
	var streamInterceptors []grpc.StreamServerInterceptor
	if telemetryStreamServerInterceptor != nil {
		streamInterceptors = append(streamInterceptors, telemetryStreamServerInterceptor)
	}
	if streamInterceptor != nil {
		streamInterceptors = append(streamInterceptors, streamInterceptor)
	}

	l = len(streamInterceptors)
	if l == 1 {
		additionalServerOptions = append(additionalServerOptions, grpc.StreamInterceptor(streamInterceptors[0]))
	} else if l > 1 {
		additionalServerOptions = append(additionalServerOptions, grpc.ChainStreamInterceptor(streamInterceptors...))
	}

	return additionalServerOptions

}

// SanitizeAsField removes initial slash if one exists and returns the new value
func SanitizeAsField(as string) string {
	if strings.HasPrefix(as, "/") {
		newas := strings.TrimPrefix(as, "/")
		return newas
	}
	return as
}
