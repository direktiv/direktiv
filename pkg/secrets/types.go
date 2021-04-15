package secrets

import (
	secretsgrpc "github.com/vorteil/direktiv/pkg/secrets/grpc"
	"google.golang.org/grpc"
)

// Backend implementations
const (
	BackendDB    = "db"
	BackendVault = "vault"

	configFile = "DIREKTIV_SECRETS_CONFIG"
)

// Server serves backend implementation
type Server struct {
	secretsgrpc.UnimplementedSecretsServiceServer
	lifeLine chan bool
	grpc     *grpc.Server

	handler secretsHandler
}

type secretsHandler interface {
	AddSecret(namespace, name string, secret []byte) error
	RemoveSecret(namespace, name string) error
	RemoveSecrets(namespace string) error
	GetSecret(namespace, name string) ([]byte, error)
	GetSecrets(namespace string) ([]string, error)
}

type dbConfig struct {
	DB  string
	Key string
}
