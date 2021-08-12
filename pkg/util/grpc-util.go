package util

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
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

	grpcSettingsFile = "/etc/direktiv/grpc-config.yaml"
)

// GrpcConfig holds the information about the grpc clients and servers
type GrpcConfig struct {
	MaxSendClient int `yaml:"max-send-client"`
	MaxRcvClient  int `yaml:"max-rcv-client"`
	MaxSendServer int `yaml:"max-send-server"`
	MaxRcvServer  int `yaml:"max-rcv-server"`

	FunctionsEndpoint string `yaml:"functions-endpoint"`
	FlowEnpoint       string `yaml:"flow-enpoint"`
	IngressEndpoint   string `yaml:"ingress-endpoint"`

	FunctionsTLS  string `yaml:"functions-tls"`
	FunctionsMTLS string `yaml:"functions-mtls"`

	IngressTLS  string `yaml:"ingress-tls"`
	IngressMTLS string `yaml:"ingress-mtls"`

	FlowTLS  string `yaml:"flow-tls"`
	FlowMTLS string `yaml:"flow-mtls"`
}

var (
	additionalServerOptions []grpc.ServerOption
	additionalCallOptions   []grpc.CallOption
	grpcCfg                 GrpcConfig

	tlsComponents map[string]tlsComponent
)

// Available grpc components in direktiv
const (
	TLSSecretsComponent   = "secrets"
	TLSIngressComponent   = "ingress"
	TLSFlowComponent      = "flow"
	TLSFunctionsComponent = "functions"
	TLSHttpComponent      = "http"
)

type tlsComponent struct {
	endpoint    string
	certificate string
	tls         string
	mtls        string
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
		tls:         grpcCfg.IngressTLS,
		mtls:        grpcCfg.IngressMTLS,
	}
	tlsComponents[TLSFunctionsComponent] = tlsComponent{
		endpoint:    FunctionsEndpoint(),
		certificate: filepath.Join(certBase, TLSFunctionsComponent),
		tls:         grpcCfg.FunctionsTLS,
		mtls:        grpcCfg.FunctionsMTLS,
	}
	tlsComponents[TLSFlowComponent] = tlsComponent{
		endpoint:    FlowEndpoint(),
		certificate: filepath.Join(certBase, TLSFlowComponent),
		tls:         grpcCfg.FlowTLS,
		mtls:        grpcCfg.FlowMTLS,
	}
	tlsComponents[TLSHttpComponent] = tlsComponent{
		endpoint:    "",
		certificate: filepath.Join(certBase, TLSHttpComponent),
	}

}

// CertsForComponent return key and cert for direktiv component
func CertsForComponent(component string) (string, string, string) {

	if c, ok := tlsComponents[component]; ok {

		if _, err := os.Stat(filepath.Join(c.certificate, "tls.key")); err != nil {
			return "", "", ""
		}

		return filepath.Join(c.certificate, "tls.key"),
			filepath.Join(c.certificate, "tls.crt"), filepath.Join(c.certificate, "ca.crt")
	}

	return "", "", ""
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

func getTransport(cert, key, cacert,
	endpoint string, server bool) (credentials.TransportCredentials, error) {

	certificate, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		log.Errorf("could not load client key pair: %s", err)
		return nil, err
	}

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(cacert)
	if err != nil {
		log.Errorf("could not read ca certificate: %s", err)
		return nil, err
	}

	// Append the certificates from the CA
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Errorf("failed to append ca certs: %v", err)
		return nil, err
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		log.Errorf("can not parse endpoint url: %v", err)
		return nil, err
	}

	tlsConfig := &tls.Config{
		ServerName:   u.Hostname(),
		Certificates: []tls.Certificate{certificate},
		RootCAs:      certPool,
	}

	if server {
		tlsConfig.ClientCAs = certPool
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}

	return credentials.NewTLS(tlsConfig), nil
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

	key, cert, cacert := CertsForComponent(component)
	if c.mtls != "none" && c.mtls != "" {

		log.Infof("using mtls for %s", component)
		creds, err := getTransport(cert, key, cacert, c.endpoint, false)
		if err != nil {
			log.Errorf("could get transport: %v", err)
			return nil, err
		}

		options = append(options, grpc.WithTransportCredentials(creds))

	} else if c.tls != "none" && c.tls != "" {

		log.Infof("using tls for %s", component)
		creds, err := credentials.NewClientTLSFromFile(cacert, "")
		if err != nil {
			return nil, fmt.Errorf("could not load ca cert: %s", err)
		}
		options = append(options, grpc.WithTransportCredentials(creds))

	} else {
		options = append(options, grpc.WithInsecure())
	}

	options = append(options, grpc.WithBalancerName(roundrobin.Name))
	options = append(options, globalGRPCDialOptions...)

	log.Infof("dialing with %s", c.endpoint)

	if len(c.endpoint) == 0 {
		return nil, fmt.Errorf("endpoint value empty")
	}

	return grpc.Dial(c.endpoint, options...)

}

// IsolateEndpoint return grpc encpoint for isolate services
func FunctionsEndpoint() string {
	return grpcCfg.FunctionsEndpoint
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
		grpcCfg.FunctionsEndpoint = os.Getenv(DirektivFunctionsEndpoint)
		grpcCfg.IngressEndpoint = os.Getenv(DirektivIngressEndpoint)

		fmt.Sscan(os.Getenv(DirektivMaxClientRcv), &grpcCfg.MaxRcvClient)
		fmt.Sscan(os.Getenv(DirektivMaxServerRcv), &grpcCfg.MaxRcvServer)
		fmt.Sscan(os.Getenv(DirektivMaxClientSend), &grpcCfg.MaxSendClient)
		fmt.Sscan(os.Getenv(DirektivMaxServerSend), &grpcCfg.MaxSendServer)

		grpcCfg.FlowTLS = os.Getenv(DirektivFlowTLS)
		grpcCfg.FlowMTLS = os.Getenv(DirektivFlowMTLS)

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

	var (
		c  tlsComponent
		ok bool
	)

	if len(bind) == 0 {
		return fmt.Errorf("grpc bind for %s empty", name)
	}

	if c, ok = tlsComponents[name]; !ok {
		return fmt.Errorf("unknown component: %s", name)
	}

	log.Debugf("%s endpoint starting at %s", name, bind)

	// use tls/mtls
	key, cert, cacert := CertsForComponent(name)
	if c.mtls != "none" && c.mtls != "" {

		log.Infof("enabling mtls for grpc service %s", name)

		creds, err := getTransport(cert, key, cacert, c.endpoint, true)
		if err != nil {
			log.Errorf("can not create grpc server: %v", err)
			return err
		}

		additionalServerOptions = append(additionalServerOptions, grpc.Creds(creds))

	} else if c.tls != "none" && c.tls != "" {

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
