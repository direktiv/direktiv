package direktiv

import (
	"context"

	"github.com/vorteil/direktiv/pkg/ingress"
	secretsgrpc "github.com/vorteil/direktiv/pkg/secrets/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (is *ingressServer) DeleteSecret(ctx context.Context, in *ingress.DeleteSecretRequest) (*emptypb.Empty, error) {

	namespace := in.GetNamespace()
	name := in.GetName()

	_, err := is.secretsClient.DeleteSecret(ctx, &secretsgrpc.SecretsDeleteRequest{
		Namespace: &namespace,
		Name:      &name,
	})

	return &emptypb.Empty{}, err

}

func (is *ingressServer) fetchSecrets(ctx context.Context, ns string) (*secretsgrpc.GetSecretsResponse, error) {

	return is.secretsClient.GetSecrets(ctx, &secretsgrpc.GetSecretsRequest{
		Namespace: &ns,
	})

}

func (is *ingressServer) GetSecrets(ctx context.Context, in *ingress.GetSecretsRequest) (*ingress.GetSecretsResponse, error) {

	output, err := is.fetchSecrets(ctx, in.GetNamespace())
	if err != nil {
		return nil, err
	}

	resp := new(ingress.GetSecretsResponse)
	for i := range output.Secrets {
		resp.Secrets = append(resp.Secrets, &ingress.GetSecretsResponse_Secret{
			Name: output.Secrets[i].Name,
		})
	}

	return resp, nil

}

type storeEncryptedRequest interface {
	GetNamespace() string
	GetName() string
	GetData() []byte
}

func (is *ingressServer) StoreSecret(ctx context.Context, in *ingress.StoreSecretRequest) (*emptypb.Empty, error) {
	var resp emptypb.Empty

	namespace := in.GetNamespace()
	name := in.GetName()

	_, err := is.secretsClient.StoreSecret(ctx, &secretsgrpc.SecretsStoreRequest{
		Namespace: &namespace,
		Name:      &name,
		Data:      in.GetData(),
	})

	return &resp, err
}
