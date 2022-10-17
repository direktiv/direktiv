package secrets

import (
	"context"
	"fmt"
	"path/filepath"
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

	if isFolder(in.GetName()) {
		return &resp, fmt.Errorf("secret required, but got folder")
	}

	if in.GetName() == "" || in.GetNamespace() == "" || len(in.GetData()) == 0 {
		return &resp, fmt.Errorf("name, namespace and secret values are required")
	}

	n := in.GetName()
	if ok := util.MatchesVarSNameAndSFName(n); !ok {
		return &resp, fmt.Errorf("secret name must match the regex pattern `%s`", util.VarSecretNameAndSecretsFolderNamePattern)
	}
	n = strings.TrimPrefix(n, "/")

	err := s.handler.AddSecret(in.GetNamespace(), n, in.GetData(), false)
	if err != nil {
		return &resp, err
	}

	if strings.Contains(n, "/") { //if is not a folder but contains / means is a secret inside a folder
		highestFolderPath := filepath.Dir(n) + "/"
		splittedPath := strings.SplitAfter(highestFolderPath, "/")
		splittedPath = splittedPath[:len(splittedPath)-1] // delete last elem cause is empty
		concPath := ""
		emptyData := []byte{}
		for _, name := range splittedPath {
			concPath += name
			err = s.handler.AddSecret(in.GetNamespace(), concPath, emptyData, true) // ignore error if is already inside
		}

	}

	return &resp, err

}

// RetrieveSecret retrieves secret from backend
func (s *Server) RetrieveSecret(ctx context.Context, in *secretsgrpc.SecretsRetrieveRequest) (*secretsgrpc.SecretsRetrieveResponse, error) {

	var resp secretsgrpc.SecretsRetrieveResponse

	if isFolder(in.GetName()) {
		return &resp, fmt.Errorf("secret name required, but got folder name")
	}

	if in.GetNamespace() == "" {
		return &resp, fmt.Errorf("namespace value are required")
	}

	n := in.GetName()
	n = strings.TrimPrefix(n, "/")
	data, err := s.handler.GetSecret(in.GetNamespace(), n)
	resp.Data = data

	return &resp, err
}

// GetSecrets returns secrets for one namespace in specific fodler
func (s *Server) GetSecrets(ctx context.Context, in *secretsgrpc.GetSecretsRequest) (*secretsgrpc.GetSecretsResponse, error) {

	var (
		resp secretsgrpc.GetSecretsResponse
		ls   []*secretsgrpc.GetSecretsResponse_Secret
	)

	if in.GetNamespace() == "" {
		return &resp, fmt.Errorf("namespace value is required")
	}

	names, err := s.handler.GetSecrets(in.GetNamespace(), in.GetName())
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

	if isFolder(in.GetName()) {
		return &resp, fmt.Errorf("secret name required, but got folder name")
	}

	if in.GetNamespace() == "" {
		return &resp, fmt.Errorf("namespace value is required")
	}

	return &resp, s.handler.RemoveSecret(in.GetNamespace(), in.GetName())
}

// DeleteNamespaceSecrets deletes secrets for a namespace
func (s *Server) DeleteNamespaceSecrets(ctx context.Context, in *secretsgrpc.DeleteNamespaceSecretsRequest) (*empty.Empty, error) {

	var resp emptypb.Empty
	return &resp, s.handler.RemoveNamespaceSecrets(in.GetNamespace())

}

// CreateFolder stores folders and create all missing folders in the path
func (s *Server) CreateFolder(ctx context.Context, in *secretsgrpc.CreateFolderRequest) (*empty.Empty, error) {

	var resp emptypb.Empty

	if !isFolder(in.GetName()) {
		return &resp, fmt.Errorf("folder name must ends with / ")
	}

	n := in.GetName()
	if ok := util.MatchesVarSNameAndSFName(n); !ok {
		return &resp, fmt.Errorf("folder name must match the regex pattern `%s`", util.VarSecretNameAndSecretsFolderNamePattern)
	}

	emptyData := []byte{}
	err := s.handler.AddSecret(in.GetNamespace(), n, emptyData, false)

	if err != nil {
		return &resp, fmt.Errorf("folder already exists")
	}

	splittedPath := strings.SplitAfter(n, "/")
	splittedPath = splittedPath[:len(splittedPath)-2] // delete last elem cause is empty and last path name cause already checked
	concPath := ""
	for _, name := range splittedPath {
		concPath += name
		err = s.handler.AddSecret(in.GetNamespace(), concPath, emptyData, false)
	}

	return &resp, err

}

// DeleteFolder deletes folder from backend
func (s *Server) DeleteFolder(ctx context.Context, in *secretsgrpc.DeleteFolderRequest) (*empty.Empty, error) {

	var resp emptypb.Empty

	if !isFolder(in.GetName()) {
		return &resp, fmt.Errorf("folder name must ends with /")
	}

	if in.GetNamespace() == "" {
		return &resp, fmt.Errorf("namespace value are required")
	}

	return &resp, s.handler.RemoveFolder(in.GetNamespace(), in.GetName())
}

func (s *Server) SearchSecret(ctx context.Context, in *secretsgrpc.SearchSecretRequest) (*secretsgrpc.SearchSecretResponse, error) {

	var (
		resp secretsgrpc.SearchSecretResponse
		ls   []*secretsgrpc.SearchSecretResponse_Secret
	)

	if in.GetNamespace() == "" {
		return &resp, fmt.Errorf("namespace value is required")
	}

	if in.GetName() == "" {
		return &resp, fmt.Errorf("name value is required")
	}

	names, err := s.handler.SearchForName(in.GetNamespace(), in.GetName())
	if err != nil {
		return &resp, err
	}

	for _, n := range names {
		var name = n
		ls = append(ls, &secretsgrpc.SearchSecretResponse_Secret{
			Name: &name,
		})
	}

	resp.Secrets = ls

	return &resp, nil

}

// StoreSecret stores secrets in backends
func (s *Server) UpdateSecret(ctx context.Context, in *secretsgrpc.UpdateSecretRequest) (*empty.Empty, error) {

	var resp emptypb.Empty

	if isFolder(in.GetName()) {
		return &resp, fmt.Errorf("secret required, but got folder")
	}

	if in.GetName() == "" || in.GetNamespace() == "" || len(in.GetData()) == 0 {
		return &resp, fmt.Errorf("name, namespace and secret values are required")
	}

	n := in.GetName()
	if ok := util.MatchesVarSNameAndSFName(n); !ok {
		return &resp, fmt.Errorf("secret name must match the regex pattern `%s`", util.VarSecretNameAndSecretsFolderNamePattern)
	}
	n = strings.TrimPrefix(n, "/")

	err := s.handler.UpdateSecret(in.GetNamespace(), n, in.GetData())
	if err != nil {
		return &resp, err
	}

	return &resp, err

}

// IsFolder Checks if name is folder
func isFolder(name string) bool {
	return (strings.HasSuffix(name, "/") || name == "")
}
