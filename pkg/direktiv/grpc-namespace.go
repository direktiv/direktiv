package direktiv

import (
	"context"
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/pkg/ingress"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (is *ingressServer) AddNamespace(ctx context.Context, in *ingress.AddNamespaceRequest) (*ingress.AddNamespaceResponse, error) {

	// TODO: can go to ent
	var resp ingress.AddNamespaceResponse
	var name string
	name = in.GetName()
	regex := "^[a-z][a-z0-9._-]{1,34}[a-z0-9]$"

	matched, err := regexp.MatchString(regex, name)
	if err != nil {
		log.Errorf("%v", NewInternalError(err))
		return nil, grpcErrInternal
	}

	if !matched {
		return nil, status.Errorf(codes.InvalidArgument, "namespace name must match regex: %s", regex)
	}

	namespace, err := is.wfServer.dbManager.addNamespace(ctx, name)
	if err != nil {
		return nil, grpcDatabaseError(err, "namespace", name)
	}

	log.Debugf("Added namespace: %v", name)

	resp.Name = &name
	resp.CreatedAt = timestamppb.New(namespace.Created)

	return &resp, nil

}

func (is *ingressServer) DeleteNamespace(ctx context.Context, in *ingress.DeleteNamespaceRequest) (*ingress.DeleteNamespaceResponse, error) {

	var resp ingress.DeleteNamespaceResponse
	var name string
	name = in.GetName()

	err := is.wfServer.dbManager.deleteNamespace(ctx, name)
	if err != nil {
		return nil, grpcDatabaseError(err, "namespace", name)
	}

	log.Debugf("Deleted namespace: %v", name)

	resp.Name = &name

	return &resp, nil

}

func (is *ingressServer) GetNamespaceLogs(ctx context.Context, in *ingress.GetNamespaceLogsRequest) (*ingress.GetNamespaceLogsResponse, error) {

	var resp ingress.GetNamespaceLogsResponse

	namespace := in.GetNamespace()
	offset := in.GetOffset()
	limit := in.GetLimit()

	logs, err := is.wfServer.instanceLogger.QueryNamespaceLogs(ctx, namespace, int(limit), int(offset))
	if err != nil {
		return nil, grpcDatabaseError(err, "namespace-logs", namespace)
	}

	resp.Offset = &offset
	resp.Limit = &limit

	for i := range logs.Logs {

		l := &logs.Logs[i]

		resp.NamespaceLogs = append(resp.NamespaceLogs, &ingress.GetNamespaceLogsResponse_NamespaceLog{
			Timestamp: timestamppb.New(time.Unix(0, l.Timestamp)),
			Message:   &l.Message,
			Context:   l.Context,
		})

	}

	return &resp, nil

}

func (is *ingressServer) GetNamespaces(ctx context.Context, in *ingress.GetNamespacesRequest) (*ingress.GetNamespacesResponse, error) {

	var resp ingress.GetNamespacesResponse
	offset := in.GetOffset()
	limit := in.GetLimit()

	namespaces, err := is.wfServer.dbManager.getNamespaces(ctx, int(offset), int(limit))
	if err != nil {
		return nil, grpcDatabaseError(err, "namespace", "")
	}

	resp.Offset = &offset
	resp.Limit = &limit

	for _, namespace := range namespaces {

		name := namespace.ID
		createdAt := namespace.Created

		resp.Namespaces = append(resp.Namespaces, &ingress.GetNamespacesResponse_Namespace{
			Name:      &name,
			CreatedAt: timestamppb.New(createdAt),
		})

	}

	return &resp, nil

}
