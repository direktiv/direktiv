package direktiv

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	ingressComponent string = "ingress"
	flowComponent    string = "flow"
	secretsComponent string = "secrets"
	healthComponent  string = "health"
)

func getEndpointTLS(config *Config, component, endpoint string) (*grpc.ClientConn, error) {

	creds, err := credentials.NewClientTLSFromFile("/etc/certs/direktiv/tls.crt", "")
	if err != nil {
		return nil, fmt.Errorf("could not load tls cert: %s", err)
	}

	return grpc.Dial(endpoint, grpc.WithTransportCredentials(creds))

}
