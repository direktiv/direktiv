package flow

import (
	"context"

	"github.com/direktiv/direktiv/pkg/flow/ent"
	entlog "github.com/direktiv/direktiv/pkg/flow/ent/logmsg"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
)

var logsOrderings = []*orderingInfo{
	{
		db:           entlog.FieldT,
		req:          "TIMESTAMP",
		defaultOrder: ent.Asc,
	},
}

var logsFilters = map[*filteringInfo]func(query *ent.LogMsgQuery, v string) (*ent.LogMsgQuery, error){}

func (flow *flow) ServerLogs(ctx context.Context, req *grpc.ServerLogsRequest) (*grpc.ServerLogsResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	query := flow.db.LogMsg.Query()
	query = query.Where(entlog.Not(entlog.HasNamespace()), entlog.Not(entlog.HasWorkflow()))

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logsFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.ServerLogsResponse)
	resp.PageInfo = pi

	err = atob(results, &resp.Results)
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

	query := flow.db.LogMsg.Query()
	query = query.Where(entlog.Not(entlog.HasNamespace()), entlog.Not(entlog.HasWorkflow()))

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logsFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.ServerLogsResponse)
	resp.PageInfo = pi

	err = atob(results, &resp.Results)
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

func (flow *flow) NamespaceLogs(ctx context.Context, req *grpc.NamespaceLogsRequest) (*grpc.NamespaceLogsResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.getNamespace(ctx, flow.db.Namespace, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	query := ns.QueryLogs()

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logsFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.NamespaceLogsResponse)
	resp.Namespace = ns.Name
	resp.PageInfo = pi

	err = atob(results, &resp.Results)
	if err != nil {
		return nil, err
	}

	return resp, nil

}

func (flow *flow) NamespaceLogsParcels(req *grpc.NamespaceLogsRequest, srv grpc.Flow_NamespaceLogsParcelsServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	var tailing bool

	ns, err := flow.getNamespace(ctx, flow.db.Namespace, req.GetNamespace())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeNamespaceLogs(ns)
	defer flow.cleanup(sub.Close)

resend:

	query := ns.QueryLogs()

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logsFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.NamespaceLogsResponse)
	resp.Namespace = ns.Name
	resp.PageInfo = pi

	err = atob(results, &resp.Results)
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

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	d, err := flow.traverseToWorkflow(ctx, flow.db.Namespace, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	query := d.wf.QueryLogs()

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logsFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.WorkflowLogsResponse)
	resp.Namespace = d.namespace()
	resp.Path = d.path
	resp.PageInfo = pi

	err = atob(results, &resp.Results)
	if err != nil {
		return nil, err
	}

	return resp, nil

}

func (flow *flow) WorkflowLogsParcels(req *grpc.WorkflowLogsRequest, srv grpc.Flow_WorkflowLogsParcelsServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	var tailing bool

	d, err := flow.traverseToWorkflow(ctx, flow.db.Namespace, req.GetNamespace(), req.GetPath())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeWorkflowLogs(d.wf)
	defer flow.cleanup(sub.Close)

resend:

	query := d.wf.QueryLogs()

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logsFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.WorkflowLogsResponse)
	resp.Namespace = d.namespace()
	resp.Path = d.path
	resp.PageInfo = pi

	err = atob(results, &resp.Results)
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

	d, err := flow.getInstance(ctx, flow.db.Namespace, req.GetNamespace(), req.GetInstance(), false)
	if err != nil {
		return nil, err
	}

	query := d.in.QueryLogs()

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logsFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.InstanceLogsResponse)
	resp.Namespace = d.namespace()
	resp.Instance = d.in.ID.String()
	resp.PageInfo = pi

	err = atob(results, &resp.Results)
	if err != nil {
		return nil, err
	}

	return resp, nil

}

func (flow *flow) InstanceLogsParcels(req *grpc.InstanceLogsRequest, srv grpc.Flow_InstanceLogsParcelsServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	var tailing bool

	d, err := flow.getInstance(ctx, flow.db.Namespace, req.GetNamespace(), req.GetInstance(), false)
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeInstanceLogs(d.in)
	defer flow.cleanup(sub.Close)

resend:

	query := d.in.QueryLogs()

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logsFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.InstanceLogsResponse)
	resp.Namespace = d.namespace()
	resp.Instance = d.in.ID.String()
	resp.PageInfo = pi

	err = atob(results, &resp.Results)
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
