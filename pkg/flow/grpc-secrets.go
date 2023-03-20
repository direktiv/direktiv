package flow

import (
	"context"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	secretsgrpc "github.com/direktiv/direktiv/pkg/secrets/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (flow *flow) Secrets(ctx context.Context, req *grpc.SecretsRequest) (*grpc.SecretsResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached := new(database.CacheData)

	err := flow.database.NamespaceByName(ctx, cached, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	p, err := getPagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	namespace := cached.Namespace.ID.String()
	name := req.GetKey()

	request := &secretsgrpc.GetSecretsRequest{
		Namespace: &namespace,
		Name:      &name,
	}

	response, err := flow.secrets.client.GetSecrets(ctx, request)
	if err != nil {
		return nil, err
	}

	cpds := newCustomPaginationDataSecrets()
	pagination := newCustomPagination(cpds)
	for i := range response.Secrets {
		cpds.Add(response.Secrets[i].GetName())
	}

	cx, err := pagination.Paginate(p)
	if err != nil {
		return nil, err
	}

	var resp grpc.SecretsResponse

	resp.Namespace = cached.Namespace.Name
	resp.Secrets = new(grpc.Secrets)
	resp.Secrets.PageInfo = new(grpc.PageInfo)

	err = bytedata.ConvertDataForOutput(cx, &resp.Secrets)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (flow *flow) SecretsStream(req *grpc.SecretsRequest, srv grpc.Flow_SecretsStreamServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	cached := new(database.CacheData)

	err := flow.database.NamespaceByName(ctx, cached, req.GetNamespace())
	if err != nil {
		return err
	}

	p, err := getPagination(req.Pagination)
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeNamespaceSecrets(cached.Namespace)
	defer flow.cleanup(sub.Close)

resend:

	namespace := cached.Namespace.ID.String()
	name := req.GetKey()

	request := &secretsgrpc.GetSecretsRequest{
		Namespace: &namespace,
		Name:      &name,
	}

	response, err := flow.secrets.client.GetSecrets(ctx, request)
	if err != nil {
		return err
	}

	cpds := newCustomPaginationDataSecrets()
	pagination := newCustomPagination(cpds)
	for i := range response.Secrets {
		cpds.Add(response.Secrets[i].GetName())
	}

	cx, err := pagination.Paginate(p)
	if err != nil {
		return err
	}

	resp := new(grpc.SecretsResponse)

	resp.Namespace = cached.Namespace.Name
	resp.Secrets = new(grpc.Secrets)
	resp.Secrets.PageInfo = new(grpc.PageInfo)

	err = bytedata.ConvertDataForOutput(cx, &resp.Secrets)
	if err != nil {
		return err
	}

	nhash = bytedata.Checksum(resp)
	if nhash != phash {
		err = srv.Send(resp)
		if err != nil {
			return err
		}
	}
	phash = nhash

	more := sub.Wait(ctx)
	if !more {
		return nil
	}

	goto resend
}

func (flow *flow) SearchSecret(ctx context.Context, req *grpc.SearchSecretRequest) (*grpc.SearchSecretResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached := new(database.CacheData)

	err := flow.database.NamespaceByName(ctx, cached, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	p, err := getPagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	namespace := cached.Namespace.ID.String()
	name := req.GetKey()

	request := &secretsgrpc.SearchSecretRequest{
		Namespace: &namespace,
		Name:      &name,
	}

	response, err := flow.secrets.client.SearchSecret(ctx, request)
	if err != nil {
		return nil, err
	}

	cpds := newCustomPaginationDataSecrets()
	pagination := newCustomPagination(cpds)
	for i := range response.Secrets {
		cpds.Add(response.Secrets[i].GetName())
	}

	cx, err := pagination.Paginate(p)
	if err != nil {
		return nil, err
	}

	var resp grpc.SearchSecretResponse

	resp.Namespace = cached.Namespace.Name
	resp.Secrets = new(grpc.Secrets)
	resp.Secrets.PageInfo = new(grpc.PageInfo)

	err = bytedata.ConvertDataForOutput(cx, &resp.Secrets)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (flow *flow) SetSecret(ctx context.Context, req *grpc.SetSecretRequest) (*grpc.SetSecretResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached := new(database.CacheData)

	err := flow.database.NamespaceByName(ctx, cached, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	namespace := cached.Namespace.ID.String()
	name := req.GetKey()

	request := &secretsgrpc.SecretsStoreRequest{
		Namespace: &namespace,
		Name:      &name,
		Data:      req.GetData(),
	}

	_, err = flow.secrets.client.StoreSecret(ctx, request)
	if err != nil {
		return nil, err
	}

	flow.logger.Infof(ctx, cached.Namespace.ID, cached.GetAttributes("namespace"), "Created namespace secret '%s'.", req.GetKey())
	flow.pubsub.NotifyNamespaceSecrets(cached.Namespace)

	var resp grpc.SetSecretResponse

	resp.Namespace = cached.Namespace.Name
	resp.Key = req.GetKey()

	return &resp, nil
}

func (flow *flow) CreateSecretsFolder(ctx context.Context, req *grpc.CreateSecretsFolderRequest) (*grpc.CreateSecretsFolderResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached := new(database.CacheData)

	err := flow.database.NamespaceByName(ctx, cached, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	namespace := cached.Namespace.ID.String()
	name := req.GetKey()

	request := &secretsgrpc.CreateSecretsFolderRequest{
		Namespace: &namespace,
		Name:      &name,
	}

	_, err = flow.secrets.client.CreateSecretsFolder(ctx, request)
	if err != nil {
		return nil, err
	}

	flow.logger.Infof(ctx, cached.Namespace.ID, cached.GetAttributes("namespace"), "Created secrets folder '%s'.", req.GetKey())
	flow.pubsub.NotifyNamespaceSecrets(cached.Namespace)

	var resp grpc.CreateSecretsFolderResponse

	resp.Namespace = cached.Namespace.Name
	resp.Key = req.GetKey()

	return &resp, nil
}

func (flow *flow) DeleteSecret(ctx context.Context, req *grpc.DeleteSecretRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached := new(database.CacheData)

	err := flow.database.NamespaceByName(ctx, cached, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	namespace := cached.Namespace.ID.String()
	name := req.GetKey()

	request := &secretsgrpc.SecretsDeleteRequest{
		Namespace: &namespace,
		Name:      &name,
	}

	_, err = flow.secrets.client.DeleteSecret(ctx, request)
	if err != nil {
		return nil, err
	}

	flow.logger.Infof(ctx, cached.Namespace.ID, cached.GetAttributes("namespace"), "Deleted namespace secret '%s'.", req.GetKey())
	flow.pubsub.NotifyNamespaceSecrets(cached.Namespace)

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) DeleteSecretsFolder(ctx context.Context, req *grpc.DeleteSecretsFolderRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached := new(database.CacheData)

	err := flow.database.NamespaceByName(ctx, cached, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	namespace := cached.Namespace.ID.String()
	name := req.GetKey()

	request := &secretsgrpc.DeleteSecretsFolderRequest{
		Namespace: &namespace,
		Name:      &name,
	}

	_, err = flow.secrets.client.DeleteSecretsFolder(ctx, request)
	if err != nil {
		return nil, err
	}

	flow.logger.Infof(ctx, cached.Namespace.ID, cached.GetAttributes("namespace"), "Deleted namespace folder '%s'.", req.GetKey())
	flow.pubsub.NotifyNamespaceSecrets(cached.Namespace)

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) UpdateSecret(ctx context.Context, req *grpc.UpdateSecretRequest) (*grpc.UpdateSecretResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached := new(database.CacheData)

	err := flow.database.NamespaceByName(ctx, cached, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	namespace := cached.Namespace.ID.String()
	name := req.GetKey()

	request := &secretsgrpc.UpdateSecretRequest{
		Namespace: &namespace,
		Name:      &name,
		Data:      req.GetData(),
	}

	_, err = flow.secrets.client.UpdateSecret(ctx, request)
	if err != nil {
		return nil, err
	}

	flow.logger.Infof(ctx, cached.Namespace.ID, cached.GetAttributes("namespace"), "Updated namespace secret '%s'.", req.GetKey())
	flow.pubsub.NotifyNamespaceSecrets(cached.Namespace)

	var resp grpc.UpdateSecretResponse

	resp.Namespace = cached.Namespace.Name
	resp.Key = req.GetKey()

	return &resp, nil
}
