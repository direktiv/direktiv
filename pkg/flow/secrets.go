package flow

import (
	"context"

	"github.com/direktiv/direktiv/pkg/flow/database"
	secretsgrpc "github.com/direktiv/direktiv/pkg/secrets/grpc"
	"github.com/direktiv/direktiv/pkg/util"
	libgrpc "google.golang.org/grpc"
)

type secrets struct {
	conn   *libgrpc.ClientConn
	client secretsgrpc.SecretsServiceClient
}

func initSecrets() (*secrets, error) {
	secrets := new(secrets)

	var err error

	secrets.conn, err = util.GetEndpointTLS("localhost:2610")
	if err != nil {
		return nil, err
	}

	secrets.client = secretsgrpc.NewSecretsServiceClient(secrets.conn)

	return secrets, nil
}

func (secrets *secrets) Close() error {
	if secrets.conn != nil {
		err := secrets.conn.Close()
		if err != nil {
			return err
		}

		secrets.conn = nil
	}

	return nil
}

func (srv *server) deleteNamespaceSecrets(ns *database.Namespace) {
	err := srv.secrets.deleteNamespaceSecrets(ns)
	if err != nil {
		srv.sugar.Error(err)
	}
}

func (secrets *secrets) deleteNamespaceSecrets(ns *database.Namespace) error {
	namespace := ns.ID.String()

	request := &secretsgrpc.DeleteNamespaceSecretsRequest{
		Namespace: &namespace,
	}

	_, err := secrets.client.DeleteNamespaceSecrets(context.Background(), request)
	if err != nil {
		return err
	}

	return nil
}
