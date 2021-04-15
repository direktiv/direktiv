package direktiv

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	secretsgrpc "github.com/vorteil/direktiv/pkg/secrets/grpc"
)

func getSecretsForInstance(ctx context.Context, instance *workflowLogicInstance, ns, name string) ([]byte, error) {

	var resp *secretsgrpc.SecretsRetrieveResponse

	resp, err := instance.engine.secretsClient.RetrieveSecret(ctx, &secretsgrpc.SecretsRetrieveRequest{
		Namespace: &ns,
		Name:      &name,
	})
	if err != nil {
		s := status.Convert(err)
		if s.Code() == codes.NotFound {
			return nil, NewUncatchableError("direktiv.secrets.notFound", "secret '%s' not found", name)
		}
		return nil, NewInternalError(err)
	}

	return resp.GetData(), nil

}
