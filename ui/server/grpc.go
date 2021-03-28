package main

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/vorteil/direktiv/pkg/ingress"
)

type grpcClient struct {
	addr   string
	client ingress.DirektivIngressClient
	tlsCfg *tls.Config
	json   jsonpb.Marshaler
}

func (g *grpcClient) init() error {

	g.initTLSConfig()

	err := g.initTLSConn()
	if err != nil {
		return err
	}

	var opts grpc.DialOption
	if insecure {
		opts = grpc.WithInsecure()
	} else {
		opts = grpc.WithTransportCredentials(credentials.NewTLS(g.tlsCfg))
	}

	cc, err := grpc.Dial(g.addr, opts)
	if err != nil {
		return err
	}

	g.client = ingress.NewDirektivIngressClient(cc)

	g.json = jsonpb.Marshaler{
		EmitDefaults: true,
	}
	return nil
}

func (g *grpcClient) initTLSConfig() {
	g.tlsCfg = &tls.Config{
		Rand:               rand.Reader,
		InsecureSkipVerify: insecure,
	}
}

func (g *grpcClient) initTLSConn() error {

	if direktivCertsDir != "" {
		keyPath := filepath.Join(direktivCertsDir, "ingress.key")
		certPath := filepath.Join(direktivCertsDir, "ingress.crt")
		caPath := filepath.Join(direktivCertsDir, "ca.pem")

		if _, err := os.Stat(keyPath); err == nil {
			certs, err := tls.LoadX509KeyPair(certPath, keyPath)
			if err != nil {
				return err
			}
			g.tlsCfg.Certificates = []tls.Certificate{certs}
		}

		if _, err := os.Stat(caPath); err == nil {

			pool := x509.NewCertPool()
			pem, err := ioutil.ReadFile(caPath)
			if err != nil {
				return err
			}

			if !pool.AppendCertsFromPEM(pem) {
				err := fmt.Errorf("can not append cert to CA")
				return err
			}

			g.tlsCfg.ClientAuth = tls.RequireAndVerifyClientCert
			g.tlsCfg.ClientCAs = pool
		}
	}

	return nil
}
