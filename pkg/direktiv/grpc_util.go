package direktiv

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

const (
	ingressComponent string = "ingress"

	serverType = "server"
	clientType = "client"
)

func tlsForGRPC(certDir, component, name string, insecure bool) (*tls.Config, error) {

	config := &tls.Config{
		Rand:               rand.Reader,
		InsecureSkipVerify: insecure,
	}

	_, err := os.Stat(certDir)
	if len(certDir) == 0 || os.IsNotExist(err) {
		return config, nil
	}

	log.Debugf("!!!!!!!!!!!!!!!!!!!checking certs in %s\n", filepath.Join(certDir, component))

	keyPath := filepath.Join(certDir, component, fmt.Sprintf("%s.key", name))
	certPath := filepath.Join(certDir, component, fmt.Sprintf("%s.pem", name))
	caPath := filepath.Join(certDir, component, "ca.pem")

	if _, err := os.Stat(keyPath); err == nil {
		log.Debugf("creating key pair")
		certs, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			log.Errorf("can not create key pair: %v\n", err)
			return nil, err
		}
		config.Certificates = []tls.Certificate{certs}
	}

	if _, err := os.Stat(caPath); err == nil {

		log.Debugf("creating ca")
		pool := x509.NewCertPool()
		pem, err := ioutil.ReadFile(caPath)
		if err != nil {
			fmt.Printf("can not create CA: %v\n", err)
			return nil, err
		}

		if !pool.AppendCertsFromPEM(pem) {
			error := fmt.Errorf("can not append cert to CA")
			log.Errorf(error.Error())
			return nil, error
		}

		config.ClientAuth = tls.RequireAndVerifyClientCert
		config.ClientCAs = pool

	}

	return config, nil

}
