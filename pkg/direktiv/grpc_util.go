package direktiv

import (
	"fmt"
	"net"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/resolver"
)

const (
	ingressComponent string = "ingress"
	flowComponent    string = "flow"
	secretsComponent string = "secrets"
	healthComponent  string = "health"

	grpcRecvMsgSizeClient = "GRPC_MAX_SEND_SIZE_CLIENT"
	grpcSendMsgSizeClient = "GRPC_MAX_RECEIVE_SIZE_CLIENT"
	grpcRecvMsgSizeServer = "GRPC_MAX_SEND_SIZE_SERVER"
	grpcSendMsgSizeServer = "GRPC_MAX_RECEIVE_SIZE_SEVER"

	// TLSCert cert
	TLSCert = "/etc/certs/direktiv/tls.crt"
	// TLSKey key
	TLSKey = "/etc/certs/direktiv/tls.key"
	// TLSCA cert CA
	TLSCA = "/etc/certs/direktiv/ca.crt"
)

func init() {
	resolver.Register(NewBuilder())
}

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
	var sizeOptions []grpc.CallOption

	// Get GRPC limits
	if maxSendOpt, ok := loadInt(grpcRecvMsgSizeClient); ok {
		log.Infof("LOADED GRPC_MAX_SEND_SIZE_CLIENT - Setting MaxCallSendMsgSize to %v", maxSendOpt)
		sizeOptions = append(sizeOptions, grpc.MaxCallSendMsgSize(maxSendOpt))
	}

	if maxRecveOpt, ok := loadInt(grpcRecvMsgSizeClient); ok {
		log.Infof("LOADED GRPC_MAX_RECEIVE_SIZE_CLIENT - Setting MaxCallRecvMsgSize to %v", maxRecveOpt)
		sizeOptions = append(sizeOptions, grpc.MaxCallRecvMsgSize(maxRecveOpt))
	}

	if len(sizeOptions) > 0 {
		options = append(options, grpc.WithDefaultCallOptions(sizeOptions...))
	}

	if _, err := os.Stat(TLSCert); !os.IsNotExist(err) {
		log.Infof("loading cert for grpc")
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

	// Get GRPC limits
	if maxSendOpt, ok := loadInt(grpcRecvMsgSizeServer); ok {
		log.Infof("LOADED GRPC_MAX_SEND_SIZE_SERVER - Setting MaxSendMsgSize to %v", maxSendOpt)
		options = append(options, grpc.MaxSendMsgSize(maxSendOpt))
	}

	if maxRecveOpt, ok := loadInt(grpcRecvMsgSizeServer); ok {
		log.Infof("LOADED GRPC_MAX_RECEIVE_SIZE_SERVER - Setting MaxRecvMsgSize to %v", maxRecveOpt)
		options = append(options, grpc.MaxRecvMsgSize(maxRecveOpt))
	}

	// Create the TLS credentials
	if _, err := os.Stat(TLSKey); !os.IsNotExist(err) {
		log.Infof("enabling tls for %s", name)
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

func loadInt(env string) (int, bool) {
	v := os.Getenv(env)
	if len(v) > 0 {
		i, err := strconv.Atoi(v)
		if err != nil {
			return 0, false
		}
		return i, true
	}

	return 0, false
}
