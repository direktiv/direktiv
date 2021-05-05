package direktiv

import (
	"fmt"
	"net"
	"os"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials"
)

const (
	ingressComponent string = "ingress"
	flowComponent    string = "flow"
	secretsComponent string = "secrets"
	healthComponent  string = "health"

	// TLSCert cert
	TLSCert = "/etc/certs/direktiv/tls.crt"
	// TLSKey key
	TLSKey = "/etc/certs/direktiv/tls.key"
)

var globalGRPCDialOptions []grpc.DialOption

func AddGlobalGRPCDialOption(opt grpc.DialOption) {
	globalGRPCDialOptions = append(globalGRPCDialOptions, opt)
}

var globalGRPCServerOptions []grpc.ServerOption

func AddGlobalGRPCServerOption(opt grpc.ServerOption) {
	globalGRPCServerOptions = append(globalGRPCServerOptions, opt)
}

// GetEndpointTLS creates a grpc client
func GetEndpointTLS(endpoint string, rr bool) (*grpc.ClientConn, error) {

	var options []grpc.DialOption

	if _, err := os.Stat(TLSCert); !os.IsNotExist(err) {
		creds, err := credentials.NewClientTLSFromFile(TLSCert, "")
		if err != nil {
			return nil, fmt.Errorf("could not load tls cert: %s", err)
		}
		options = append(options, grpc.WithTransportCredentials(creds))
	} else {
		options = append(options, grpc.WithInsecure())
	}

	if rr {
		options = append(options, grpc.WithBalancerName(roundrobin.Name))
	}

	options = append(options, globalGRPCDialOptions...)

	return grpc.Dial(endpoint, options...)

}

// GrpcStart starts a grpc server
func GrpcStart(server **grpc.Server, name, bind string, register func(srv *grpc.Server)) error {

	log.Debugf("%s endpoint starting at %s", name, bind)

	var options []grpc.ServerOption

	// Create the TLS credentials
	if _, err := os.Stat(TLSKey); !os.IsNotExist(err) {
		creds, err := credentials.NewServerTLSFromFile(TLSCert, TLSKey)
		if err != nil {
			return fmt.Errorf("could not load TLS keys: %s", err)
		}
		options = append(options, grpc.Creds(creds))
	}

	listener, err := net.Listen("tcp", bind)
	if err != nil {
		return err
	}

	options = append(options, globalGRPCServerOptions...)

	(*server) = grpc.NewServer(options...)

	register(*server)

	go (*server).Serve(listener)

	return nil

}
