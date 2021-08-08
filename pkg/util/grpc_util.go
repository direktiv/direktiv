package util

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/resolver"
	"gopkg.in/yaml.v2"
)

// GRPC constants
const (
	IngressComponent string = "ingress"
	FlowComponent    string = "flow"

	// TLSCert cert
	TLSCert = "/etc/certs/direktiv/tls.crt"
	// TLSKey key
	TLSKey = "/etc/certs/direktiv/tls.key"
	// TLSCA cert CA
	TLSCA = "/etc/certs/direktiv/ca.crt"

	grpcSettingsFile = "/etc/direktiv/grpc-config.yaml"
)

type grpConfig struct {
	MaxSendClient int `yaml:"max-send-client"`
	MaxRcvClient  int `yaml:"max-rcv-client"`
	MaxSendServer int `yaml:"max-send-server"`
	MaxRcvServer  int `yaml:"max-rcv-server"`

	IsolateEndpoint string `yaml:"isolate-endpoint"`
	FlowEnpoint     string `yaml:"flow-enpoint"`
	IngressEndpoint string `yaml:"ingress-endpoint"`
}

var (
	additionalServerOptions []grpc.ServerOption
	additionalCallOptions   []grpc.CallOption
	gcfg                    grpConfig
)

func init() {
	resolver.Register(NewBuilder())
}

var (
	globalGRPCDialOptions []grpc.DialOption
)

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

	if len(additionalCallOptions) > 0 {
		options = append(options,
			grpc.WithDefaultCallOptions(additionalCallOptions...))
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

// IsolateEndpoint return grpc encpoint for isolate services
func IsolateEndpoint() string {
	return gcfg.IsolateEndpoint
}

// IngressEndpoint return grpc encpoint for ingress services
func IngressEndpoint() string {
	return gcfg.IngressEndpoint
}

// FlowEndpoint return grpc encpoint for flow services
func FlowEndpoint() string {
	return gcfg.FlowEnpoint
}

// GRPCUnmarshalConfig reads default grpc config file in
func GRPCUnmarshalConfig() {

	cfgBytes, err := ioutil.ReadFile(grpcSettingsFile)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(cfgBytes, &gcfg)
	if err != nil {
		return
	}
	log.Infof("setting grpc server send/rcv size: %v/%v", gcfg.MaxSendServer, gcfg.MaxRcvServer)
	additionalServerOptions = append(additionalServerOptions, grpc.MaxSendMsgSize(gcfg.MaxSendServer))
	additionalServerOptions = append(additionalServerOptions, grpc.MaxRecvMsgSize(gcfg.MaxRcvServer))
	log.Infof("setting grpc client send/rcv size: %v/%v", gcfg.MaxSendClient, gcfg.MaxRcvClient)
	additionalCallOptions = append(additionalCallOptions, grpc.MaxCallSendMsgSize(gcfg.MaxSendClient))
	additionalCallOptions = append(additionalCallOptions, grpc.MaxCallRecvMsgSize(gcfg.MaxRcvClient))

}

// GrpcStart starts a grpc server
func GrpcStart(server **grpc.Server, name, bind string, register func(srv *grpc.Server)) error {

	if len(bind) == 0 {
		return fmt.Errorf("grpc bind for %s empty", name)
	}

	log.Debugf("%s endpoint starting at %s", name, bind)

	// Create the TLS credentials
	if _, err := os.Stat(TLSKey); !os.IsNotExist(err) {
		log.Infof("enabling tls for %s", name)
		creds, err := credentials.NewServerTLSFromFile(TLSCert, TLSKey)
		if err != nil {
			return fmt.Errorf("could not load TLS keys: %s", err)
		}
		additionalServerOptions = append(additionalServerOptions, grpc.Creds(creds))
	}

	listener, err := net.Listen("tcp", bind)
	if err != nil {
		return err
	}

	additionalServerOptions = append(additionalServerOptions, globalGRPCServerOptions...)

	(*server) = grpc.NewServer(additionalServerOptions...)

	register(*server)

	go (*server).Serve(listener)

	return nil

}
