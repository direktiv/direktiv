package flow

import (
	"context"
	"errors"
	"strings"

	"github.com/direktiv/direktiv/pkg/flow/ent"
	entinst "github.com/direktiv/direktiv/pkg/flow/ent/instance"
	entlog "github.com/direktiv/direktiv/pkg/flow/ent/logmsg"
	entlogtag "github.com/direktiv/direktiv/pkg/flow/ent/logtag"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/google/uuid"
)

var logsOrderings = []*orderingInfo{
	{
		db:           entlog.FieldT,
		req:          "TIMESTAMP",
		defaultOrder: ent.Asc,
	},
}

var logsFilters = map[*filteringInfo]func(query *ent.LogMsgQuery, v string) (*ent.LogMsgQuery, error){
	{
		field: "ID",
		ftype: "MATCH",
	}: func(query *ent.LogMsgQuery, v string) (*ent.LogMsgQuery, error) {
		id, err := uuid.Parse(v)
		if err != nil {
			return nil, err
		}
		return query.Where(entlog.HasInstanceWith(entinst.IDEQ(id))).WithLogtag(), nil
	},
	{
		field: "STATE",
		ftype: "MATCH",
	}: func(query *ent.LogMsgQuery, v string) (*ent.LogMsgQuery, error) {
		return query.Where(entlog.HasLogtagWith(entlogtag.And(entlogtag.Type("state"), entlogtag.Value(v)))), nil
	},
	{
		field: "QUERY",
		ftype: "MATCH",
	}: func(query *ent.LogMsgQuery, v string) (*ent.LogMsgQuery, error) {
		values := strings.Split(v, "::")
		if len(values) != 3 {
			return nil, errors.New("wrong argument number use iterator::wfname::state")
		}
		query = query.Where(entlog.HasLogtagWith(entlogtag.And(entlogtag.Type("iterator"), entlogtag.Value(values[0]))))
		query = query.Where(entlog.HasLogtagWith(entlogtag.And(entlogtag.Type("name"), entlogtag.Value(values[1]))))
		return query.Where(entlog.HasLogtagWith(entlogtag.And(entlogtag.Type("state"), entlogtag.Value(values[2])))), nil
	},
}

func (flow *flow) ServerLogs(ctx context.Context, req *grpc.ServerLogsRequest) (*grpc.ServerLogsResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	query := flow.db.LogMsg.Query().WithLogtag()
	query = query.Where(entlog.Not(entlog.HasNamespace()), entlog.Not(entlog.HasWorkflow()))

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logsFilters)
	if err != nil {
		return nil, err
	}
	var respTag []map[string]string
	for i := 0; i < len(results); i++ {
		tags := make(map[string]string)
		for _, tag := range results[i].Edges.Logtag {
			tags[tag.Type] = tag.Value
		}
		respTag = append(respTag, tags)
	}
	resp := new(grpc.ServerLogsResponse)
	resp.PageInfo = pi

	err = atob(results, &resp.Results)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(resp.Results); i++ {
		resp.Results[i].Tags = respTag[i]
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

	query := flow.db.LogMsg.Query().WithLogtag()
	query = query.Where(entlog.Not(entlog.HasNamespace()), entlog.Not(entlog.HasWorkflow()))

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logsFilters)
	if err != nil {
		return err
	}
	var respTag []map[string]string
	for i := 0; i < len(results); i++ {
		tags := make(map[string]string)
		for _, tag := range results[i].Edges.Logtag {
			tags[tag.Type] = tag.Value
		}
		respTag = append(respTag, tags)
	}
	resp := new(grpc.ServerLogsResponse)
	resp.PageInfo = pi

	err = atob(results, &resp.Results)
	if err != nil {
		return err
	}
	for i := 0; i < len(resp.Results); i++ {
		resp.Results[i].Tags = respTag[i]
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

	query := ns.QueryLogs().WithLogtag()

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logsFilters)
	if err != nil {
		return nil, err
	}
	var respTag []map[string]string
	for i := 0; i < len(results); i++ {
		tags := make(map[string]string)
		for _, tag := range results[i].Edges.Logtag {
			tags[tag.Type] = tag.Value
		}
		respTag = append(respTag, tags)
	}
	resp := new(grpc.NamespaceLogsResponse)
	resp.Namespace = ns.Name
	resp.PageInfo = pi

	err = atob(results, &resp.Results)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(resp.Results); i++ {
		resp.Results[i].Tags = respTag[i]
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

	query := ns.QueryLogs().WithLogtag()

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logsFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.NamespaceLogsResponse)
	resp.Namespace = ns.Name
	resp.PageInfo = pi
	var respTag []map[string]string
	for i := 0; i < len(results); i++ {
		tags := make(map[string]string)
		for _, tag := range results[i].Edges.Logtag {
			tags[tag.Type] = tag.Value
		}
		respTag = append(respTag, tags)
	}
	err = atob(results, &resp.Results)
	if err != nil {
		return err
	}
	for i := 0; i < len(resp.Results); i++ {
		resp.Results[i].Tags = respTag[i]
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

	query := d.wf.QueryLogs().WithLogtag()

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logsFilters)
	if err != nil {
		return nil, err
	}
	var respTag []map[string]string
	for i := 0; i < len(results); i++ {
		tags := make(map[string]string)
		for _, tag := range results[i].Edges.Logtag {
			tags[tag.Type] = tag.Value
		}
		respTag = append(respTag, tags)
	}
	resp := new(grpc.WorkflowLogsResponse)
	resp.Namespace = d.namespace()
	resp.Path = d.path
	resp.PageInfo = pi

	err = atob(results, &resp.Results)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(resp.Results); i++ {
		resp.Results[i].Tags = respTag[i]
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

	query := d.wf.QueryLogs().WithLogtag()

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logsFilters)
	if err != nil {
		return err
	}
	var respTag []map[string]string
	for i := 0; i < len(results); i++ {
		tags := make(map[string]string)
		for _, tag := range results[i].Edges.Logtag {
			tags[tag.Type] = tag.Value
		}
		respTag = append(respTag, tags)
	}
	resp := new(grpc.WorkflowLogsResponse)
	resp.Namespace = d.namespace()
	resp.Path = d.path
	resp.PageInfo = pi

	err = atob(results, &resp.Results)
	if err != nil {
		return err
	}
	for i := 0; i < len(resp.Results); i++ {
		resp.Results[i].Tags = respTag[i]
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
	// uuids :=
	// for _, v := range d.in.Parents {
	// uid, err := uuid.Parse(v)
	// if err != nil {
	// 	return nil, err
	// }
	query := flow.db.Instance.QueryLogn(d.in).WithLogtag()
	//query := d.in.QueryLogs().WithLogtag()
	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logsFilters)
	if err != nil {
		return nil, err
	}
	var respTag []map[string]string
	for i := 0; i < len(results); i++ {
		tags := make(map[string]string)
		for _, tag := range results[i].Edges.Logtag {
			tags[tag.Type] = tag.Value
		}
		respTag = append(respTag, tags)
	}
	resp := new(grpc.InstanceLogsResponse)
	resp.Namespace = d.namespace()
	resp.Instance = d.in.ID.String()
	resp.PageInfo = pi

	err = atob(results, &resp.Results)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(resp.Results); i++ {
		resp.Results[i].Tags = respTag[i]
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

	//query := d.in.QueryLogs().WithLogtag()
	query := flow.db.Instance.QueryLogn(d.in).WithLogtag()
	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logsFilters)
	if err != nil {
		return err
	}
	resp := new(grpc.InstanceLogsResponse)
	resp.Namespace = d.namespace()
	resp.Instance = d.in.ID.String()
	resp.PageInfo = pi
	var respTag []map[string]string
	for i := 0; i < len(results); i++ {
		tags := make(map[string]string)
		for _, tag := range results[i].Edges.Logtag {
			tags[tag.Type] = tag.Value
		}
		respTag = append(respTag, tags)
	}
	err = atob(results, &resp.Results)

	if err != nil {
		return err
	}
	for i := 0; i < len(resp.Results); i++ {
		resp.Results[i].Tags = respTag[i]
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
