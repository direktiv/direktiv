package flow

import (
	"context"

	libgrpc "google.golang.org/grpc"

	"github.com/vorteil/direktiv/pkg/flow/ent"
	secretsgrpc "github.com/vorteil/direktiv/pkg/secrets/grpc"
	"github.com/vorteil/direktiv/pkg/util"
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

func (srv *server) deleteNamespaceSecrets(ns *ent.Namespace) {

	err := srv.secrets.deleteNamespaceSecrets(ns)
	if err != nil {
		srv.sugar.Error(err)
	}

}

func (secrets *secrets) deleteNamespaceSecrets(ns *ent.Namespace) error {

	namespace := ns.ID.String()

	request := &secretsgrpc.DeleteSecretsRequest{
		Namespace: &namespace,
	}

	_, err := secrets.client.DeleteSecrets(context.Background(), request)
	if err != nil {
		return err
	}

	return nil

}
