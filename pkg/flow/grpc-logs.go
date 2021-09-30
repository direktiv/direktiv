package flow

import (
	"context"

	"github.com/vorteil/direktiv/pkg/flow/ent"
	entlog "github.com/vorteil/direktiv/pkg/flow/ent/logmsg"
	"github.com/vorteil/direktiv/pkg/flow/grpc"
)

func (flow *flow) ServerLogs(ctx context.Context, req *grpc.ServerLogsRequest) (*grpc.ServerLogsResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	p, err := getPagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	opts := []ent.LogMsgPaginateOption{}
	opts = append(opts, logsOrder(p))
	filter := logsFilter(p)
	if filter != nil {
		opts = append(opts, filter)
	}

	logc := flow.db.LogMsg
	query := logc.Query()
	query = query.Where(entlog.Not(entlog.HasNamespace()), entlog.Not(entlog.HasWorkflow()))
	cx, err := query.Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return nil, err
	}

	var resp grpc.ServerLogsResponse

	err = atob(cx, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil

}

func (flow *flow) ServerLogsParcels(req *grpc.ServerLogsRequest, srv grpc.Flow_ServerLogsParcelsServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	var tailing bool

	p, err := getPagination(req.Pagination)
	if err != nil {
		return err
	}

	porder := p.order
	pfilter := p.filter

	opts := []ent.LogMsgPaginateOption{}
	opts = append(opts, logsOrder(p))
	filter := logsFilter(p)
	if filter != nil {
		opts = append(opts, filter)
	}

	sub := flow.pubsub.SubscribeServerLogs()
	defer flow.cleanup(sub.Close)

resend:

	logc := flow.db.LogMsg
	query := logc.Query()
	query = query.Where(entlog.Not(entlog.HasNamespace()), entlog.Not(entlog.HasWorkflow()))
	cx, err := query.Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return err
	}

	var resp = new(grpc.ServerLogsResponse)

	err = atob(cx, resp)
	if err != nil {
		return err
	}

	if len(resp.Edges) != 0 || !tailing {

		tailing = true

		err = srv.Send(resp)
		if err != nil {
			return err
		}

		p = new(pagination)
		p.after = resp.PageInfo.EndCursor
		p.order = porder
		p.filter = pfilter

	}

	more := sub.Wait(ctx)
	if !more {
		return nil
	}

	goto resend

}

func (flow *flow) NamespaceLogs(ctx context.Context, req *grpc.NamespaceLogsRequest) (*grpc.NamespaceLogsResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	p, err := getPagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	opts := []ent.LogMsgPaginateOption{}
	opts = append(opts, logsOrder(p))
	filter := logsFilter(p)
	if filter != nil {
		opts = append(opts, filter)
	}

	nsc := flow.db.Namespace

	namespace := req.GetNamespace()

	ns, err := flow.getNamespace(ctx, nsc, namespace)
	if err != nil {
		return nil, err
	}

	cx, err := ns.QueryLogs().Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return nil, err
	}

	var resp grpc.NamespaceLogsResponse

	err = atob(cx, &resp)
	if err != nil {
		return nil, err
	}

	resp.Namespace = namespace

	return &resp, nil

}

func (flow *flow) NamespaceLogsParcels(req *grpc.NamespaceLogsRequest, srv grpc.Flow_NamespaceLogsParcelsServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	var tailing bool

	p, err := getPagination(req.Pagination)
	if err != nil {
		return err
	}

	porder := p.order
	pfilter := p.filter

	opts := []ent.LogMsgPaginateOption{}
	opts = append(opts, logsOrder(p))
	filter := logsFilter(p)
	if filter != nil {
		opts = append(opts, filter)
	}

	nsc := flow.db.Namespace

	namespace := req.GetNamespace()

	ns, err := flow.getNamespace(ctx, nsc, namespace)
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeNamespaceLogs(ns)
	defer flow.cleanup(sub.Close)

resend:

	cx, err := ns.QueryLogs().Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return err
	}

	var resp = new(grpc.NamespaceLogsResponse)

	err = atob(cx, resp)
	if err != nil {
		return err
	}

	resp.Namespace = namespace

	if len(resp.Edges) != 0 || !tailing {

		tailing = true

		err = srv.Send(resp)
		if err != nil {
			return err
		}

		p = new(pagination)
		p.after = resp.PageInfo.EndCursor
		p.order = porder
		p.filter = pfilter

	}

	more := sub.Wait(ctx)
	if !more {
		return nil
	}

	goto resend

}

func (flow *flow) WorkflowLogs(ctx context.Context, req *grpc.WorkflowLogsRequest) (*grpc.WorkflowLogsResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	p, err := getPagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	opts := []ent.LogMsgPaginateOption{}
	opts = append(opts, logsOrder(p))
	filter := logsFilter(p)
	if filter != nil {
		opts = append(opts, filter)
	}

	nsc := flow.db.Namespace
	d, err := flow.traverseToWorkflow(ctx, nsc, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	cx, err := d.wf.QueryLogs().Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return nil, err
	}

	var resp grpc.WorkflowLogsResponse

	err = atob(cx, &resp)
	if err != nil {
		return nil, err
	}

	resp.Namespace = d.namespace()
	resp.Path = d.path

	return &resp, nil

}

func (flow *flow) WorkflowLogsParcels(req *grpc.WorkflowLogsRequest, srv grpc.Flow_WorkflowLogsParcelsServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	var tailing bool

	p, err := getPagination(req.Pagination)
	if err != nil {
		return err
	}

	porder := p.order
	pfilter := p.filter

	opts := []ent.LogMsgPaginateOption{}
	opts = append(opts, logsOrder(p))
	filter := logsFilter(p)
	if filter != nil {
		opts = append(opts, filter)
	}

	nsc := flow.db.Namespace
	d, err := flow.traverseToWorkflow(ctx, nsc, req.GetNamespace(), req.GetPath())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeWorkflowLogs(d.wf)
	defer flow.cleanup(sub.Close)

resend:

	cx, err := d.wf.QueryLogs().Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return err
	}

	var resp = new(grpc.WorkflowLogsResponse)

	err = atob(cx, resp)
	if err != nil {
		return err
	}

	resp.Namespace = d.namespace()
	resp.Path = d.path

	if len(resp.Edges) != 0 || !tailing {

		tailing = true

		err = srv.Send(resp)
		if err != nil {
			return err
		}

		p = new(pagination)
		p.after = resp.PageInfo.EndCursor
		p.order = porder
		p.filter = pfilter

	}

	more := sub.Wait(ctx)
	if !more {
		return nil
	}

	goto resend

}

func (flow *flow) InstanceLogs(ctx context.Context, req *grpc.InstanceLogsRequest) (*grpc.InstanceLogsResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	p, err := getPagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	opts := []ent.LogMsgPaginateOption{}
	opts = append(opts, logsOrder(p))
	filter := logsFilter(p)
	if filter != nil {
		opts = append(opts, filter)
	}

	nsc := flow.db.Namespace
	d, err := flow.getInstance(ctx, nsc, req.GetNamespace(), req.GetInstance(), false)
	if err != nil {
		return nil, err
	}

	cx, err := d.in.QueryLogs().Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return nil, err
	}

	var resp grpc.InstanceLogsResponse

	err = atob(cx, &resp)
	if err != nil {
		return nil, err
	}

	resp.Namespace = d.namespace()
	resp.Instance = d.in.ID.String()

	return &resp, nil

}

func (flow *flow) InstanceLogsParcels(req *grpc.InstanceLogsRequest, srv grpc.Flow_InstanceLogsParcelsServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	var tailing bool

	p, err := getPagination(req.Pagination)
	if err != nil {
		return err
	}

	porder := p.order
	pfilter := p.filter

	opts := []ent.LogMsgPaginateOption{}
	opts = append(opts, logsOrder(p))
	filter := logsFilter(p)
	if filter != nil {
		opts = append(opts, filter)
	}

	nsc := flow.db.Namespace
	d, err := flow.getInstance(ctx, nsc, req.GetNamespace(), req.GetInstance(), false)
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeInstanceLogs(d.in)
	defer flow.cleanup(sub.Close)

resend:

	cx, err := d.in.QueryLogs().Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return err
	}

	var resp = new(grpc.InstanceLogsResponse)

	err = atob(cx, resp)
	if err != nil {
		return err
	}

	resp.Namespace = d.namespace()
	resp.Instance = d.in.ID.String()

	if len(resp.Edges) != 0 || !tailing {

		tailing = true

		err = srv.Send(resp)
		if err != nil {
			return err
		}

		p = new(pagination)
		p.after = resp.PageInfo.EndCursor
		p.order = porder
		p.filter = pfilter

	}

	more := sub.Wait(ctx)
	if !more {
		return nil
	}

	goto resend

}
