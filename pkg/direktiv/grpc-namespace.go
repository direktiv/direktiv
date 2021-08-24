package direktiv

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/vorteil/direktiv/pkg/ingress"
	"github.com/vorteil/direktiv/pkg/util"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (is *ingressServer) AddNamespace(ctx context.Context, in *ingress.AddNamespaceRequest) (*ingress.AddNamespaceResponse, error) {

	// TODO: can go to ent
	var resp ingress.AddNamespaceResponse
	var name string
	name = in.GetName()
	if ok := util.MatchesRegex(name); !ok {
		return nil, fmt.Errorf("namespace name must comply with the regex pattern `%s`", util.RegexPattern)
	}

	namespace, err := is.wfServer.dbManager.addNamespace(ctx, name)
	if err != nil {
		return nil, grpcDatabaseError(err, "namespace", name)
	}

	appLog.Debugf("Added namespace: %v", name)

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

	appLog.Debugf("Deleted namespace: %v", name)

	// delete all functions
	err = deleteKnativeFunctions(is.wfServer.engine.functionsClient, in.GetName(), "", "")
	if err != nil {
		appLog.Errorf("can not delete knative functions: %v", err)
	}

	resp.Name = &name

	return &resp, nil

}

func (is *ingressServer) GetNamespaceLogs(ctx context.Context, in *ingress.GetNamespaceLogsRequest) (*ingress.GetNamespaceLogsResponse, error) {

	var resp ingress.GetNamespaceLogsResponse

	namespace := in.GetNamespace()
	offset := in.GetOffset()
	limit := in.GetLimit()

	lc := is.wfServer.components[util.LogComponent].(*logClient)
	r, err := lc.logsForNamespace(namespace, offset, limit)
	if err != nil {
		return nil, grpcDatabaseError(err, "namespace-logs", namespace)
	}

	for i := range r {
		infoMap := r[i]

		// get msg
		msg := infoMap["msg"].(string)

		// get sec
		ts := infoMap["ts"].(float64)
		sec, dec := math.Modf(ts)

		resp.NamespaceLogs = append(resp.NamespaceLogs, &ingress.GetNamespaceLogsResponse_NamespaceLog{
			Message:   &msg,
			Timestamp: timestamppb.New(time.Unix(int64(sec), int64(dec*(1e9)))),
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
