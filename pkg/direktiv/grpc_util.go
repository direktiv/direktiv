package direktiv

import (
	"fmt"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	ingressComponent string = "ingress"
	flowComponent    string = "flow"
	secretsComponent string = "secrets"
	healthComponent  string = "health"

	// TLSCert/TLSKey files, ampped in as secrets
	TLSCert = "/etc/certs/direktiv/tls.crt"
	TLSKey  = "/etc/certs/direktiv/tls.key"
)

func getEndpointTLS(config *Config, component, endpoint string) (*grpc.ClientConn, error) {

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

	return grpc.Dial(endpoint, options...)

}
