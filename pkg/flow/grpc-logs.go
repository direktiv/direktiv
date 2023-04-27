package flow

import (
	"context"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/flow/internallogger"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
)

type fileAttributes filestore.File

func (f fileAttributes) GetAttributes() map[string]string {
	m := make(map[string]string)
	m["namespace-id"] = f.RootID.String()
	if f.Typ == filestore.FileTypeWorkflow {
		m["workflow-id"] = f.ID.String()
	}
	return m
}

func (flow *flow) ServerLogs(ctx context.Context, req *grpc.ServerLogsRequest) (*grpc.ServerLogsResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ql := internallogger.QueryLogs()
	ql.WhereWorkflowIsNil()
	ql.WhereNamespaceIsNIl()
	ql.WhereInstanceIsNIl()
	logs, err := ql.GetAll(ctx, flow.gormDB)
	if err != nil {
		return nil, err
	}
	pi := BuildPageInfo(ql)

	resp := new(grpc.ServerLogsResponse)
	resp.PageInfo = &pi

	resp.Results, err = bytedata.ConvertLogMsgToGrpcLog(logs)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (flow *flow) ServerLogsParcels(req *grpc.ServerLogsRequest, srv grpc.Flow_ServerLogsParcelsServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	var tailing bool

	sub := flow.pubsub.SubscribeServerLogs()
	defer flow.cleanup(sub.Close)

resend:
	ql := internallogger.QueryLogs()
	ql.WhereWorkflowIsNil()
	ql.WhereNamespaceIsNIl()
	ql.WhereInstanceIsNIl()
	logs, err := ql.GetAll(ctx, flow.gormDB)
	if err != nil {
		return err
	}
	ql.WithLimit(int(req.Pagination.Limit))
	ql.WithOffset(int(req.Pagination.Limit))
	pi := BuildPageInfo(ql)

	resp := new(grpc.ServerLogsResponse)
	resp.Results, err = bytedata.ConvertLogMsgToGrpcLog(logs)
	if err != nil {
		return err
	}
	resp.PageInfo = &pi
	if len(resp.Results) != 0 || !tailing {
		tailing = true

		err = srv.Send(resp)
		if err != nil {
			return err
		}

		req.Pagination.Offset += int32(len(resp.Results))
	}

	more := sub.Wait(ctx)
	if !more {
		return nil
	}
	goto resend
}

func (flow *flow) NamespaceLogs(ctx context.Context, req *grpc.NamespaceLogsRequest) (*grpc.NamespaceLogsResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())
	cached := new(database.CacheData)

	err := flow.database.NamespaceByName(ctx, cached, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	ql := internallogger.QueryLogs()
	id := cached.Namespace.ID
	ql.WhereNamespace(id)
	logs, err := ql.GetAll(ctx, flow.gormDB)
	if err != nil {
		return nil, err
	}
	pi := BuildPageInfo(ql)

	resp := new(grpc.NamespaceLogsResponse)
	resp.Namespace = req.Namespace
	resp.PageInfo = &pi

	resp.Results, err = bytedata.ConvertLogMsgToGrpcLog(logs)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (flow *flow) NamespaceLogsParcels(req *grpc.NamespaceLogsRequest, srv grpc.Flow_NamespaceLogsParcelsServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	var tailing bool
	cached := new(database.CacheData)

	err := flow.database.NamespaceByName(ctx, cached, req.GetNamespace())
	if err != nil {
		return err
	}
	id := cached.Namespace.ID

	sub := flow.pubsub.SubscribeNamespaceLogs(&id)
	defer flow.cleanup(sub.Close)

resend:
	ql := internallogger.QueryLogs()

	ql.WhereNamespace(id)
	ql.WithLimit(int(req.Pagination.Limit))
	ql.WithOffset(int(req.Pagination.Limit))
	logs, err := ql.GetAll(ctx, flow.gormDB)
	if err != nil {
		return err
	}
	pi := BuildPageInfo(ql)

	resp := new(grpc.NamespaceLogsResponse)
	resp.Namespace = req.Namespace
	resp.PageInfo = &pi

	resp.Results, err = bytedata.ConvertLogMsgToGrpcLog(logs)
	if err != nil {
		return err
	}
	if len(resp.Results) != 0 || !tailing {
		tailing = true

		err = srv.Send(resp)
		if err != nil {
			return err
		}

		req.Pagination.Offset += int32(len(resp.Results))
	}

	more := sub.Wait(ctx)
	if !more {
		return nil
	}

	goto resend
}

func (flow *flow) WorkflowLogs(ctx context.Context, req *grpc.WorkflowLogsRequest) (*grpc.WorkflowLogsResponse, error) {
	flow.sugar.Errorf("Handling gRPC request: %s", this())

	ns, f, err := flow.getWorkflow(ctx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	ql := internallogger.QueryLogs()
	id := f.ID
	ql.WhereWorkflow(id)
	logs, err := ql.GetAll(ctx, flow.gormDB)
	if err != nil {
		return nil, err
	}
	pi := grpc.PageInfo{}

	resp := new(grpc.WorkflowLogsResponse)
	resp.Namespace = ns.Name
	resp.Path = f.Path
	resp.PageInfo = &pi

	resp.Results, err = bytedata.ConvertLogMsgToGrpcLog(logs)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (flow *flow) WorkflowLogsParcels(req *grpc.WorkflowLogsRequest, srv grpc.Flow_WorkflowLogsParcelsServer) error {
	flow.sugar.Errorf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	var tailing bool

	ns, f, err := flow.getWorkflow(ctx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeWorkflowLogs(f.ID)
	defer flow.cleanup(sub.Close)

resend:

	ql := internallogger.QueryLogs()
	id := f.ID
	ql.WhereWorkflow(id)
	ql.WithLimit(int(req.Pagination.Limit))
	ql.WithOffset(int(req.Pagination.Offset))
	logs, err := ql.GetAll(ctx, flow.gormDB)
	if err != nil {
		return err
	}
	pi := BuildPageInfo(ql)

	resp := new(grpc.WorkflowLogsResponse)
	resp.Namespace = ns.Name
	resp.Path = f.Path
	resp.PageInfo = &pi

	resp.Results, err = bytedata.ConvertLogMsgToGrpcLog(logs)
	if err != nil {
		return err
	}

	if len(resp.Results) != 0 || !tailing {
		tailing = true

		err = srv.Send(resp)
		if err != nil {
			return err
		}

		req.Pagination.Offset += int32(len(resp.Results))
	}

	more := sub.Wait(ctx)
	if !more {
		return nil
	}
	goto resend
}

func (flow *flow) InstanceLogs(ctx context.Context, req *grpc.InstanceLogsRequest) (*grpc.InstanceLogsResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached, err := flow.getInstance(ctx, req.GetNamespace(), req.GetInstance())
	if err != nil {
		return nil, err
	}

	// its important to append the instanceID to the callpath since we don't do it when creating the database entry
	prefix := internallogger.AppendInstanceID(cached.Instance.CallPath, cached.Instance.ID.String())
	root, err := internallogger.GetRootinstanceID(prefix)
	if err != nil {
		return nil, err
	}

	ql := internallogger.QueryLogs()
	callerIsRoot := root == cached.Instance.Invoker

	ql.WhereRootInstanceIdEQ(root)
	if !callerIsRoot {
		ql.WhereInstanceCallPathHasPrefix(prefix)
	}

	logs, err := ql.GetAll(ctx, flow.gormDB)
	if err != nil {
		return nil, err
	}
	pi := BuildPageInfo(ql)

	results, err := buildInstanceLogResp(ctx, logs, &pi, req.Pagination, req.Namespace, req.Instance)
	if err != nil {
		return nil, err
	}

	resp := results

	return resp, nil
}

func (flow *flow) InstanceLogsParcels(req *grpc.InstanceLogsRequest, srv grpc.Flow_InstanceLogsParcelsServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	var tailing bool

	cached, err := flow.getInstance(ctx, req.GetNamespace(), req.GetInstance())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeInstanceLogs(cached)
	defer flow.cleanup(sub.Close)
	// its important to append the intanceID to the callpath since we don't do it when creating the database entry.
	prefix := internallogger.AppendInstanceID(cached.Instance.CallPath, cached.Instance.ID.String())
	root, err := internallogger.GetRootinstanceID(prefix)
	callerIsRoot := root == cached.Instance.ID.String()
	if err != nil {
		return err
	}

resend:

	ql := internallogger.QueryLogs()

	ql.WhereRootInstanceIdEQ(root)
	if !callerIsRoot {
		ql.WhereInstanceCallPathHasPrefix(prefix)
	}

	ql.WithLimit(int(req.Pagination.Limit))
	ql.WithOffset(int(req.Pagination.Offset))
	logs, err := ql.GetAll(ctx, flow.gormDB)
	if err != nil {
		return err
	}
	pi := BuildPageInfo(ql)

	results, err := buildInstanceLogResp(ctx, logs, &pi, req.Pagination, req.Namespace, req.Instance)
	if err != nil {
		return err
	}

	resp := results

	if len(resp.Results) != 0 || !tailing {
		tailing = true

		err = srv.Send(resp)
		if err != nil {
			return err
		}

		req.Pagination.Offset += int32(len(resp.Results))
	}

	more := sub.Wait(ctx)
	if !more {
		return nil
	}

	goto resend
}

func buildInstanceLogResp(ctx context.Context,
	in []*internallogger.LogMsgs,
	pi *grpc.PageInfo,
	page *grpc.Pagination,
	namespace string,
	instance string,
) (*grpc.InstanceLogsResponse, error) {
	filters := page.Filter
	results := in
	for _, v := range filters {
		results = internallogger.FilterLogmsg(v, in)
		pi.Total = int32(len(results))
	}

	resp := new(grpc.InstanceLogsResponse)
	resp.Namespace = namespace
	resp.Instance = instance
	resp.PageInfo = pi
	var err error
	resp.Results, err = bytedata.ConvertLogMsgToGrpcLog(results)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func BuildPageInfo(lq internallogger.LogMsgQuery) grpc.PageInfo {
	return grpc.PageInfo{
		Limit:  int32(lq.GetLimit()),
		Offset: int32(lq.GetOffset()),
	}
}
