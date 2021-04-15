package secrets

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/vorteil/direktiv/pkg/direktiv"
	secretsgrpc "github.com/vorteil/direktiv/pkg/secrets/grpc"
)

// NewServer creates a new secrets server
func NewServer(backend string) (*Server, error) {

	srv := &Server{
		lifeLine: make(chan bool),
	}

	log.Infof("starting secret backend %s", backend)

	switch backend {
	case BackendDB:
		err := srv.setupDB()
		if err != nil {
			return nil, err
		}
	case BackendVault:

	default:
		return nil, fmt.Errorf("backend %s not supported", backend)
	}

	return srv, nil

}

// Run starts the secrets server
func (s *Server) Run() {

	log.Infof("starting secret server")

	direktiv.GrpcStart(&s.grpc, "ingress", "127.0.0.1:2610", func(srv *grpc.Server) {
		secretsgrpc.RegisterSecretsServiceServer(srv, s)
	})

}

// Stop stops the server gracefully
func (s *Server) Stop() {

	go func() {

		log.Infof("stopping workflow server")
		s.lifeLine <- true

	}()
}

// Kill kills the server
func (s *Server) Kill() {

	go func() {

		defer func() {
			_ = recover()
		}()

		s.lifeLine <- true

	}()

}

// Lifeline interface impl
func (s *Server) Lifeline() chan bool {
	return s.lifeLine
}

// StoreSecret stroes secrets in backends
func (s *Server) StoreSecret(ctx context.Context, in *secretsgrpc.SecretsStoreRequest) (*empty.Empty, error) {

	var resp emptypb.Empty

	if in.GetName() == "" || in.GetNamespace() == "" || len(in.GetData()) == 0 {
		return &resp, fmt.Errorf("name, namespace and secret values are required")
	}

	err := s.handler.AddSecret(in.GetNamespace(), in.GetName(), in.GetData())

	return &resp, err

}

// RetrieveSecret retrieves secret from backend
func (s *Server) RetrieveSecret(ctx context.Context, in *secretsgrpc.SecretsRetrieveRequest) (*secretsgrpc.SecretsRetrieveResponse, error) {

	var resp secretsgrpc.SecretsRetrieveResponse

	if in.GetName() == "" || in.GetNamespace() == "" {
		return &resp, fmt.Errorf("name and namespace values are required")
	}

	data, err := s.handler.GetSecret(in.GetNamespace(), in.GetName())
	resp.Data = data

	return &resp, err
}

// GetSecrets returns secrets for one namespace
func (s *Server) GetSecrets(ctx context.Context, in *secretsgrpc.GetSecretsRequest) (*secretsgrpc.GetSecretsResponse, error) {

	var (
		resp secretsgrpc.GetSecretsResponse
		ls   []*secretsgrpc.GetSecretsResponse_Secret
	)

	if in.GetNamespace() == "" {
		return &resp, fmt.Errorf("namespace is required")
	}

	names, err := s.handler.GetSecrets(in.GetNamespace())
	if err != nil {
		return &resp, err
	}

	for _, n := range names {
		var name = n
		ls = append(ls, &secretsgrpc.GetSecretsResponse_Secret{
			Name: &name,
		})
	}

	resp.Secrets = ls

	return &resp, nil

}

// DeleteSecret deletes single secret from backend
func (s *Server) DeleteSecret(ctx context.Context, in *secretsgrpc.SecretsDeleteRequest) (*empty.Empty, error) {

	var resp emptypb.Empty

	if in.GetName() == "" || in.GetNamespace() == "" {
		return &resp, fmt.Errorf("name and namespace values are required")
	}

	return &resp, s.handler.RemoveSecret(in.GetNamespace(), in.GetName())
}

// DeleteSecrets deletes secrets for a namespace
func (s *Server) DeleteSecrets(ctx context.Context, in *secretsgrpc.DeleteSecretsRequest) (*empty.Empty, error) {

	var resp emptypb.Empty
	return &resp, s.handler.RemoveSecrets(in.GetNamespace())

}
