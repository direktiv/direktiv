package flow

import (
	"context"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/direktiv/direktiv/pkg/refactor/pubsub"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (flow *flow) Secrets(ctx context.Context, req *grpc.SecretsRequest) (*grpc.SecretsResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	list, err := tx.DataStore().Secrets().GetAll(ctx, ns.Name)
	if err != nil {
		return nil, err
	}

	var resp grpc.SecretsResponse

	resp.Namespace = ns.Name
	resp.Secrets = new(grpc.Secrets)
	// TODO: investigate is PageInfo can be nil.
	resp.Secrets.PageInfo = nil

	resp.Secrets.Results = bytedata.ConvertSecretsToGrpcSecretList(list)

	return &resp, nil
}

func (flow *flow) SecretsStream(req *grpc.SecretsRequest, srv grpc.Flow_SecretsStreamServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())
	ctx := srv.Context()

	resp, err := flow.Secrets(ctx, req)
	if err != nil {
		return err
	}
	// mock streaming response.
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			err = srv.Send(resp)
			if err != nil {
				return err
			}
			time.Sleep(time.Second * 5)
		}
	}
}

func (flow *flow) SearchSecret(ctx context.Context, req *grpc.SearchSecretRequest) (*grpc.SearchSecretResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	list, err := tx.DataStore().Secrets().GetAll(ctx, ns.Name)
	if err != nil {
		return nil, err
	}

	var resp grpc.SearchSecretResponse

	resp.Namespace = ns.Name
	resp.Secrets = new(grpc.Secrets)
	// TODO: investigate is PageInfo can be nil.
	resp.Secrets.PageInfo = nil

	resp.Secrets.Results = bytedata.ConvertSecretsToGrpcSecretList(list)

	return &resp, nil
}

func (flow *flow) SetSecret(ctx context.Context, req *grpc.SetSecretRequest) (*grpc.SetSecretResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	err = tx.DataStore().Secrets().Set(ctx, &datastore.Secret{
		Namespace: ns.Name,
		Name:      req.GetKey(),
		Data:      req.GetData(),
	})
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	err = flow.pBus.Publish(pubsub.SecretCreate, ns.Name)
	if err != nil {
		flow.sugar.Error("pubsub publish", "error", err)
	}

	var resp grpc.SetSecretResponse

	resp.Namespace = ns.Name
	resp.Key = req.GetKey()

	return &resp, nil
}

func (flow *flow) CreateSecretsFolder(ctx context.Context, req *grpc.CreateSecretsFolderRequest) (*grpc.CreateSecretsFolderResponse, error) {
	// TODO: ask jens if this feature still required.
	//nolint
	return nil, nil
}

func (flow *flow) DeleteSecret(ctx context.Context, req *grpc.DeleteSecretRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	err = tx.DataStore().Secrets().Delete(ctx, ns.Name, req.GetKey())
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	err = flow.pBus.Publish(pubsub.SecretDelete, ns.Name)
	if err != nil {
		flow.sugar.Error("pubsub publish", "error", err)
	}

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) DeleteSecretsFolder(ctx context.Context, req *grpc.DeleteSecretsFolderRequest) (*emptypb.Empty, error) {
	// TODO: ask jens if this feature still required.
	//nolint
	return nil, nil
}

func (flow *flow) UpdateSecret(ctx context.Context, req *grpc.UpdateSecretRequest) (*grpc.UpdateSecretResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	err = tx.DataStore().Secrets().Update(ctx, &datastore.Secret{
		Namespace: ns.Name,
		Name:      req.GetKey(),
		Data:      req.GetData(),
	})
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	err = flow.pBus.Publish(pubsub.SecretUpdate, ns.Name)
	if err != nil {
		flow.sugar.Error("pubsub publish", "error", err)
	}

	var resp grpc.UpdateSecretResponse

	resp.Namespace = ns.Name
	resp.Key = req.GetKey()

	return &resp, nil
}
