package secrets

import (
	secretsgrpc "github.com/direktiv/direktiv/pkg/secrets/grpc"
	"github.com/direktiv/direktiv/pkg/secrets/handler"
	"google.golang.org/grpc"
)

// Server serves backend implementation.
type Server struct {
	secretsgrpc.UnimplementedSecretsServiceServer
	lifeLine chan bool
	grpc     *grpc.Server

	handler handler.SecretsHandler
}
