package flow

import (
	"context"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (flow *flow) Secrets(ctx context.Context, req *grpc.SecretsRequest) (*grpc.SecretsResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	list, err := tx.DataStore().Secrets().GetAll(ctx, ns.ID)
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

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	list, err := tx.DataStore().Secrets().Search(ctx, ns.ID, req.GetKey())
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

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	err = tx.DataStore().Secrets().Set(ctx, &core.Secret{
		ID:          uuid.New(),
		NamespaceID: ns.ID,
		Name:        req.GetKey(),
		Data:        req.GetData(),
	})
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	// TODO: Alax, please look into this.
	// flow.logger.Infof(ctx, cached.Namespace.ID, cached.GetAttributes(recipient.Namespace), "Created namespace secret '%s'.", req.GetKey())
	// flow.pubsub.NotifyNamespaceSecrets(cached.Namespace)

	var resp grpc.SetSecretResponse

	resp.Namespace = ns.Name
	resp.Key = req.GetKey()

	return &resp, nil
}

func (flow *flow) CreateSecretsFolder(ctx context.Context, req *grpc.CreateSecretsFolderRequest) (*grpc.CreateSecretsFolderResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	err = tx.DataStore().Secrets().CreateFolder(ctx, ns.ID, req.GetKey())
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	// TODO: Alex please look into this.
	// flow.logger.Infof(ctx, cached.Namespace.ID, cached.GetAttributes(recipient.Namespace), "Created secrets folder '%s'.", req.GetKey())
	// flow.pubsub.NotifyNamespaceSecrets(cached.Namespace)

	var resp grpc.CreateSecretsFolderResponse

	resp.Namespace = ns.Name
	resp.Key = req.GetKey()

	return &resp, nil
}

func (flow *flow) DeleteSecret(ctx context.Context, req *grpc.DeleteSecretRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	err = tx.DataStore().Secrets().Delete(ctx, ns.ID, req.GetKey())
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	// TODO: Alex please look into this.
	// flow.logger.Infof(ctx, cached.Namespace.ID, cached.GetAttributes(recipient.Namespace), "Deleted namespace secret '%s'.", req.GetKey())
	// flow.pubsub.NotifyNamespaceSecrets(cached.Namespace)

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) DeleteSecretsFolder(ctx context.Context, req *grpc.DeleteSecretsFolderRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	err = tx.DataStore().Secrets().DeleteFolder(ctx, ns.ID, req.GetKey())
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	// TODO: Alex please look into this.
	// flow.logger.Infof(ctx, cached.Namespace.ID, cached.GetAttributes(recipient.Namespace), "Deleted namespace folder '%s'.", req.GetKey())
	// flow.pubsub.NotifyNamespaceSecrets(cached.Namespace)

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) UpdateSecret(ctx context.Context, req *grpc.UpdateSecretRequest) (*grpc.UpdateSecretResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	err = tx.DataStore().Secrets().Update(ctx, &core.Secret{
		NamespaceID: ns.ID,
		Name:        req.GetKey(),
		Data:        req.GetData(),
	})
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	// TODO: Alex, please look into this.
	// flow.logger.Infof(ctx, cached.Namespace.ID, cached.GetAttributes(recipient.Namespace), "Updated namespace secret '%s'.", req.GetKey())
	// flow.pubsub.NotifyNamespaceSecrets(cached.Namespace)

	var resp grpc.UpdateSecretResponse

	resp.Namespace = ns.Name
	resp.Key = req.GetKey()

	return &resp, nil
}
