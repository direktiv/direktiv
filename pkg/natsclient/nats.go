package natsclient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"os"

	"github.com/nats-io/nats.go"
)

func Connect() (*nats.Conn, error) {
	// set the deployment name in dns names
	deploymentName := os.Getenv("DIREKTIV_DEPLOYMENT_NAME")

	return nats.Connect(
		fmt.Sprintf("tls://%s-nats.default.svc:4222", deploymentName),
		nats.ClientTLSConfig(
			func() (tls.Certificate, error) {
				cert, err := tls.LoadX509KeyPair("/etc/direktiv-tls/server.crt",
					"/etc/direktiv-tls/server.key")
				if err != nil {
					slog.Error("cannot create certificate pair", slog.Any("error", err))
					return tls.Certificate{}, err
				}

				return cert, nil
			},
			func() (*x509.CertPool, error) {
				caCert, err := os.ReadFile("/etc/direktiv-tls/ca.crt")
				if err != nil {
					return nil, err
				}
				caPool := x509.NewCertPool()
				if !caPool.AppendCertsFromPEM(caCert) {
					slog.Error("cannot create certificate pair", slog.Any("error", err))
					return nil, err
				}

				return caPool, nil
			},
		),
	)
}
