package direktiv

import (
	"context"
	"fmt"
	"strings"

	"encoding/base64"

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

func (is *ingressServer) DeleteRegistry(ctx context.Context, in *ingress.DeleteRegistryRequest) (*emptypb.Empty, error) {
	var resp emptypb.Empty

	err := kubernetesDeleteSecret(in.GetName(), in.GetNamespace())

	return &resp, err
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

func (is *ingressServer) GetRegistries(ctx context.Context, in *ingress.GetRegistriesRequest) (*ingress.GetRegistriesResponse, error) {

	resp := new(ingress.GetRegistriesResponse)

	regs, err := kubernetesListRegistries(in.GetNamespace())

	if err != nil {
		return resp, err
	}

	for _, reg := range regs {
		split := strings.SplitN(reg, "###", 2)

		if len(split) != 2 {
			return nil, fmt.Errorf("invalid registry format")
		}

		resp.Registries = append(resp.Registries, &ingress.GetRegistriesResponse_Registry{
			Name: &split[0],
			Id:   &split[1],
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

func (is *ingressServer) StoreRegistry(ctx context.Context, in *ingress.StoreRegistryRequest) (*emptypb.Empty, error) {
	var resp emptypb.Empty

	// create secret data, needs to be attached to service account
	userToken := strings.SplitN(string(in.Data), ":", 2)
	if len(userToken) != 2 {
		return nil, fmt.Errorf("invalid username/token format")
	}

	tmpl := `{
	"auths": {
		"%s": {
			"username": "%s",
			"password": "%s",
			"auth": "%s"
		}
	}
	}`

	auth := fmt.Sprintf(tmpl, in.GetName(), userToken[0], userToken[1],
		base64.StdEncoding.EncodeToString(in.Data))

	err := kubernetesAddSecret(in.GetName(), in.GetNamespace(), []byte(auth))
	if err != nil {
		return nil, err
	}

	return &resp, nil

}
