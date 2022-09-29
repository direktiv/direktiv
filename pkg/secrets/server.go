package secrets

import (
	"context"
	"fmt"
	"strings"

	"github.com/direktiv/direktiv/pkg/dlog"
	secretsgrpc "github.com/direktiv/direktiv/pkg/secrets/grpc"
	"github.com/direktiv/direktiv/pkg/secrets/handler"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var logger *zap.SugaredLogger

// NewServer creates a new secrets server
func NewServer(backend string) (*Server, error) {

	var err error

	logger, err = dlog.ApplicationLogger("secrets")
	if err != nil {
		return nil, err
	}

	srv := &Server{
		lifeLine: make(chan bool),
	}

	if backend == "" {
		backend = "db"
	}

	logger.Infof("starting secret backend %s", backend)
	backendType, err := handler.ParseType(backend)
	if err != nil {
		return nil, err
	}

	srv.handler, err = backendType.Setup()
	if err != nil {
		return nil, err
	}

	return srv, nil

}

// Run starts the secrets server
func (s *Server) Run() {

	logger.Infof("starting secret server")

	util.GrpcStart(&s.grpc, "secrets", "127.0.0.1:2610", func(srv *grpc.Server) {
		secretsgrpc.RegisterSecretsServiceServer(srv, s)
	})

}

// Stop stops the server gracefully
func (s *Server) Stop() {

	go func() {

		logger.Infof("stopping secret server")
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

// StoreSecret stores secrets in backends
func (s *Server) StoreSecret(ctx context.Context, in *secretsgrpc.SecretsStoreRequest) (*empty.Empty, error) {

	var resp emptypb.Empty

	if strings.HasSuffix(in.GetName(), "/") || in.GetName() == "" {
		return &resp, fmt.Errorf("secret required, but got folder")
	}

	if in.GetName() == "" || in.GetNamespace() == "" || len(in.GetData()) == 0 {
		return &resp, fmt.Errorf("name, namespace and secret values are required")
	}

	n := in.GetName()
	if ok := util.MatchesVarRegex(n); !ok {
		return &resp, fmt.Errorf("secret name must match the regex pattern `%s`", util.RegexPattern)
	}
	n = strings.TrimPrefix(n, "/")

	err := s.handler.AddSecret(in.GetNamespace(), n, in.GetData())

	return &resp, err

}

// RetrieveSecret retrieves secret from backend
func (s *Server) RetrieveSecret(ctx context.Context, in *secretsgrpc.SecretsRetrieveRequest) (*secretsgrpc.SecretsRetrieveResponse, error) {

	var resp secretsgrpc.SecretsRetrieveResponse

	if strings.HasSuffix(in.GetName(), "/") {
		return &resp, fmt.Errorf("secret name required, but got folder name")
	}

	if in.GetName() == "" || in.GetNamespace() == "" {
		return &resp, fmt.Errorf("name and namespace values are required")
	}

	n := in.GetName()
	n = strings.TrimPrefix(n, "/")
	data, err := s.handler.GetSecret(in.GetNamespace(), n)
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

	if strings.HasSuffix(in.GetName(), "/") {
		return &resp, fmt.Errorf("secret name required, but got folder name")
	}

	if in.GetName() == "" || in.GetNamespace() == "" {
		return &resp, fmt.Errorf("name and namespace values are required")
	}

	return &resp, s.handler.RemoveSecret(in.GetNamespace(), in.GetName())
}

// DeleteSecrets deletes secrets for a namespace //TODO Rename toi DeleteNamespaceSecrets
func (s *Server) DeleteSecrets(ctx context.Context, in *secretsgrpc.DeleteSecretsRequest) (*empty.Empty, error) {

	var resp emptypb.Empty
	return &resp, s.handler.RemoveSecrets(in.GetNamespace())

}

// AddFolder stores folders and create all missing folders in the path
func (s *Server) AddFolder(ctx context.Context, in *secretsgrpc.SecretsStoreRequest) (*empty.Empty, error) {

	var resp emptypb.Empty

	if !strings.HasSuffix(in.GetName(), "/") {
		return &resp, fmt.Errorf("folder name must ends with / ")
	}

	if len(in.GetData()) != 0 {
		return &resp, fmt.Errorf("folder must have empty data ")
	}

	n := in.GetName()
	if ok := util.MatchesVarRegex(n); !ok {
		return &resp, fmt.Errorf("folder name must match the regex pattern `%s`", util.RegexPattern)
	}

	err := s.handler.AddSecret(in.GetNamespace(), n, in.GetData())

	if err != nil {
		return &resp, fmt.Errorf("folder already exists")
	}

	splittedPath := strings.SplitAfter(n, "/")
	splittedPath = splittedPath[:len(splittedPath)-2] // delete last elem cause is empty and last path name cause already checked
	concPath := ""
	for _, name := range splittedPath {
		concPath += name
		err = s.handler.AddSecret(in.GetNamespace(), concPath, in.GetData())
	}

	return &resp, err

}

// DeleteFolder deletes folder from backend
func (s *Server) DeleteFolder(ctx context.Context, in *secretsgrpc.SecretsDeleteRequest) (*empty.Empty, error) {

	var resp emptypb.Empty

	if !strings.HasSuffix(in.GetName(), "/") {
		return &resp, fmt.Errorf("folder name must ends with /")
	}

	if in.GetName() == "" || in.GetNamespace() == "" {
		return &resp, fmt.Errorf("name and namespace values are required")
	}

	return &resp, s.handler.RemoveSecret(in.GetNamespace(), in.GetName())
}
