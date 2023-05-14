package flow

import (
	"context"

	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (flow *flow) Secrets(ctx context.Context, req *grpc.SecretsRequest) (*grpc.SecretsResponse, error) {
	return nil, nil
}

func (flow *flow) SecretsStream(req *grpc.SecretsRequest, srv grpc.Flow_SecretsStreamServer) error {
	return nil
}

func (flow *flow) SearchSecret(ctx context.Context, req *grpc.SearchSecretRequest) (*grpc.SearchSecretResponse, error) {
	return nil, nil
}

func (flow *flow) SetSecret(ctx context.Context, req *grpc.SetSecretRequest) (*grpc.SetSecretResponse, error) {
	return nil, nil
}

func (flow *flow) CreateSecretsFolder(ctx context.Context, req *grpc.CreateSecretsFolderRequest) (*grpc.CreateSecretsFolderResponse, error) {
	return nil, nil
}

func (flow *flow) DeleteSecret(ctx context.Context, req *grpc.DeleteSecretRequest) (*emptypb.Empty, error) {
	return nil, nil
}

func (flow *flow) DeleteSecretsFolder(ctx context.Context, req *grpc.DeleteSecretsFolderRequest) (*emptypb.Empty, error) {
	return nil, nil
}

func (flow *flow) UpdateSecret(ctx context.Context, req *grpc.UpdateSecretRequest) (*grpc.UpdateSecretResponse, error) {
	return nil, nil
}
