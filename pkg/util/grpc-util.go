package util

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

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

	certBase = "/etc/direktiv/certs/"

	// TLSCert cert
	TLSCert = "/etc/certs/direktiv/tls.crt"
	// TLSKey key
	TLSKey = "/etc/certs/direktiv/tls.key"
	// TLSCA cert CA
	TLSCA = "/etc/certs/direktiv/ca.crt"

	grpcSettingsFile = "/etc/direktiv/grpc-config.yaml"
)

// GrpcConfig holds the information about the grpc clients and servers
type GrpcConfig struct {
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
	grpcCfg                 GrpcConfig

	tlsComponents map[string]tlsComponent
)

// Available grpc components in direktiv
const (
	TLSSecretsComponent  = "secrets"
	TLSIngressComponent  = "ingress"
	TLSFlowComponent     = "flow"
	TLSIsolatesComponent = "isolates"
	TLSHttpComponent     = "http"
)

type tlsComponent struct {
	endpoint    string
	certificate string
}

func init() {

	resolver.Register(NewBuilder())

	grpcUnmarshalConfig()

	tlsComponents = make(map[string]tlsComponent)

	tlsComponents[TLSSecretsComponent] = tlsComponent{
		endpoint:    "127.0.0.1:2610",
		certificate: filepath.Join(certBase, TLSSecretsComponent),
	}
	tlsComponents[TLSIngressComponent] = tlsComponent{
		endpoint:    IngressEndpoint(),
		certificate: filepath.Join(certBase, TLSIngressComponent),
	}
	tlsComponents[TLSIsolatesComponent] = tlsComponent{
		endpoint:    IsolateEndpoint(),
		certificate: filepath.Join(certBase, TLSIsolatesComponent),
	}
	tlsComponents[TLSFlowComponent] = tlsComponent{
		endpoint:    FlowEndpoint(),
		certificate: filepath.Join(certBase, TLSFlowComponent),
	}
	tlsComponents[TLSHttpComponent] = tlsComponent{
		endpoint:    "",
		certificate: filepath.Join(certBase, TLSHttpComponent),
	}

}

// CertsForComponent return key and cert for direktiv component
func CertsForComponent(component string) (string, string) {

	if c, ok := tlsComponents[component]; ok {

		if _, err := os.Stat(filepath.Join(c.certificate, "tls.key")); err != nil {
			return "", ""
		}

		return filepath.Join(c.certificate, "tls.key"),
			filepath.Join(c.certificate, "tls.crt")
	}

	return "", ""
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
func GetEndpointTLS(component string) (*grpc.ClientConn, error) {

	var (
		c  tlsComponent
		ok bool
	)

	if c, ok = tlsComponents[component]; !ok {
		return nil, fmt.Errorf("unknown component: %s", component)
	}

	var options []grpc.DialOption

	if len(additionalCallOptions) > 0 {
		options = append(options,
			grpc.WithDefaultCallOptions(additionalCallOptions...))
	}

	// if _, err := os.Stat(TLSCert); !os.IsNotExist(err) {
	// 	log.Infof("loading cert for grpc")
	// 	creds, err := credentials.NewClientTLSFromFile(TLSCert, "")
	// 	if err != nil {
	// 		return nil, fmt.Errorf("could not load tls cert: %s", err)
	// 	}
	// 	options = append(options, grpc.WithTransportCredentials(creds))
	// } else {
	options = append(options, grpc.WithInsecure())
	// }

	options = append(options, grpc.WithBalancerName(roundrobin.Name))
	options = append(options, globalGRPCDialOptions...)

	log.Infof("dialing with %s", c.endpoint)

	if len(c.endpoint) == 0 {
		return nil, fmt.Errorf("endpoint value empty")
	}

	return grpc.Dial(c.endpoint, options...)

}

// IsolateEndpoint return grpc encpoint for isolate services
func IsolateEndpoint() string {
	return grpcCfg.IsolateEndpoint
}

// IngressEndpoint return grpc encpoint for ingress services
func IngressEndpoint() string {
	return grpcCfg.IngressEndpoint
}

// FlowEndpoint return grpc encpoint for flow services
func FlowEndpoint() string {
	return grpcCfg.FlowEnpoint
}

// GrpcCfg returns the full grpc configuration
func GrpcCfg() GrpcConfig {
	return grpcCfg
}

func grpcUnmarshalConfig() {

	// try to build the grpc config from envs
	if _, err := os.Stat("/etc/direktiv/grpc-config.yaml"); os.IsNotExist(err) {

		grpcCfg.FlowEnpoint = os.Getenv(DirektivFlowEndpoint)
		grpcCfg.IsolateEndpoint = os.Getenv(DirektivIsolateEndpoint)
		grpcCfg.IngressEndpoint = os.Getenv(DirektivIngressEndpoint)

		fmt.Sscan(os.Getenv(DirektivMaxClientRcv), &grpcCfg.MaxRcvClient)
		fmt.Sscan(os.Getenv(DirektivMaxServerRcv), &grpcCfg.MaxRcvServer)
		fmt.Sscan(os.Getenv(DirektivMaxClientSend), &grpcCfg.MaxSendClient)
		fmt.Sscan(os.Getenv(DirektivMaxServerSend), &grpcCfg.MaxSendServer)

	} else {
		cfgBytes, err := ioutil.ReadFile(grpcSettingsFile)
		if err != nil {
			return
		}

		err = yaml.Unmarshal(cfgBytes, &grpcCfg)
		if err != nil {
			return
		}
	}

	log.Infof("setting grpc server send/rcv size: %v/%v", grpcCfg.MaxSendServer, grpcCfg.MaxRcvServer)
	additionalServerOptions = append(additionalServerOptions, grpc.MaxSendMsgSize(grpcCfg.MaxSendServer))
	additionalServerOptions = append(additionalServerOptions, grpc.MaxRecvMsgSize(grpcCfg.MaxRcvServer))
	log.Infof("setting grpc client send/rcv size: %v/%v", grpcCfg.MaxSendClient, grpcCfg.MaxRcvClient)
	additionalCallOptions = append(additionalCallOptions, grpc.MaxCallSendMsgSize(grpcCfg.MaxSendClient))
	additionalCallOptions = append(additionalCallOptions, grpc.MaxCallRecvMsgSize(grpcCfg.MaxRcvClient))

}

// GrpcStart starts a grpc server
func GrpcStart(server **grpc.Server, name, bind string, register func(srv *grpc.Server)) error {

	if len(bind) == 0 {
		return fmt.Errorf("grpc bind for %s empty", name)
	}

	log.Debugf("%s endpoint starting at %s", name, bind)

	// use tls if key file found
	key, cert := CertsForComponent(name)
	if len(key) > 0 {
		log.Infof("enabling tls for grpc service %s", name)
		creds, err := credentials.NewServerTLSFromFile(cert, key)
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
