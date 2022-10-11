package flow

import (
	"context"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/grpc"
	secretsgrpc "github.com/direktiv/direktiv/pkg/secrets/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (flow *flow) Secrets(ctx context.Context, req *grpc.SecretsRequest) (*grpc.SecretsResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.getNamespace(ctx, flow.db.Namespace, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	namespace := ns.ID.String()
	name := req.GetKey()

	p, err := getPagination(req.Pagination)
	if err != nil {
		return nil, err
	}

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

	resp.Namespace = ns.Name
	resp.Secrets = new(grpc.Secrets)
	resp.Secrets.PageInfo = new(grpc.PageInfo)

	err = atob(cx, &resp.Secrets)
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

	ns, err := flow.getNamespace(ctx, flow.db.Namespace, req.GetNamespace())
	if err != nil {
		return err
	}

	namespace := ns.ID.String()
	name := req.GetKey()

	p, err := getPagination(req.Pagination)
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeNamespaceSecrets(ns)
	defer flow.cleanup(sub.Close)

resend:

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

	var resp = new(grpc.SecretsResponse)

	resp.Namespace = ns.Name
	resp.Secrets = new(grpc.Secrets)
	resp.Secrets.PageInfo = new(grpc.PageInfo)

	err = atob(cx, &resp.Secrets)
	if err != nil {
		return err
	}

	nhash = checksum(resp)
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

func (flow *flow) SetSecret(ctx context.Context, req *grpc.SetSecretRequest) (*grpc.SetSecretResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.getNamespace(ctx, flow.db.Namespace, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	namespace := ns.ID.String()
	name := req.GetKey()

	request := &secretsgrpc.SecretsStoreRequest{
		Namespace: &namespace,
		Name:      &name,
		Data:      req.GetData(),
	}

	_, err = flow.secrets.client.StoreSecret(ctx, request)
	if err != nil {
		fmt.Println("==== FAILED TO STORE SECRET", namespace)
		return nil, err
	}

	flow.logToNamespace(ctx, time.Now(), ns, "Created namespace secret '%s'.", req.GetKey())
	flow.pubsub.NotifyNamespaceSecrets(ns)

	var resp grpc.SetSecretResponse

	resp.Namespace = ns.Name
	resp.Key = req.GetKey()

	return &resp, nil

}

func (flow *flow) CreateSecretsFolder(ctx context.Context, req *grpc.CreateSecretsFolderRequest) (*grpc.CreateSecretsFolderResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.getNamespace(ctx, flow.db.Namespace, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	namespace := ns.ID.String()
	name := req.GetKey()

	request := &secretsgrpc.CreateFolderRequest{
		Namespace: &namespace,
		Name:      &name,
	}

	_, err = flow.secrets.client.CreateFolder(ctx, request)

	flow.logToNamespace(ctx, time.Now(), ns, "Created secrets folder '%s'.", req.GetKey())
	flow.pubsub.NotifyNamespaceSecrets(ns)

	var resp grpc.CreateSecretsFolderResponse

	resp.Namespace = ns.Name
	resp.Key = req.GetKey()

	return &resp, nil

}

func (flow *flow) DeleteSecret(ctx context.Context, req *grpc.DeleteSecretRequest) (*emptypb.Empty, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.getNamespace(ctx, flow.db.Namespace, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	namespace := ns.ID.String()
	name := req.GetKey()

	request := &secretsgrpc.SecretsDeleteRequest{
		Namespace: &namespace,
		Name:      &name,
	}

	_, err = flow.secrets.client.DeleteSecret(ctx, request)
	if err != nil {
		return nil, err
	}

	flow.logToNamespace(ctx, time.Now(), ns, "Deleted namespace secret '%s'.", req.GetKey())
	flow.pubsub.NotifyNamespaceSecrets(ns)

	var resp emptypb.Empty

	return &resp, nil

}

func (flow *flow) DeleteFolder(ctx context.Context, req *grpc.DeleteFolderRequest) (*emptypb.Empty, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.getNamespace(ctx, flow.db.Namespace, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	namespace := ns.ID.String()
	name := req.GetKey()

	request := &secretsgrpc.DeleteFolderRequest{
		Namespace: &namespace,
		Name:      &name,
	}

	_, err = flow.secrets.client.DeleteFolder(ctx, request)
	if err != nil {
		return nil, err
	}

	flow.logToNamespace(ctx, time.Now(), ns, "Deleted namespace folder '%s'.", req.GetKey())
	flow.pubsub.NotifyNamespaceSecrets(ns)

	var resp emptypb.Empty

	return &resp, nil

}
