package flow

import (
	"context"
	"strings"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	entinst "github.com/direktiv/direktiv/pkg/flow/ent/instance"
	entlog "github.com/direktiv/direktiv/pkg/flow/ent/logmsg"
	entns "github.com/direktiv/direktiv/pkg/flow/ent/namespace"
	entwf "github.com/direktiv/direktiv/pkg/flow/ent/workflow"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/flow/internallogger"
	"github.com/google/uuid"
)

var logsOrderings = []*orderingInfo{
	{
		db:           entlog.FieldT,
		req:          "TIMESTAMP",
		defaultOrder: ent.Asc,
	},
}

var logEntFilters = map[*filteringInfo]func(query *ent.LogMsgQuery, v string) (*ent.LogMsgQuery, error){
	{
		field: "ID",
		ftype: "MATCH",
	}: func(query *ent.LogMsgQuery, v string) (*ent.LogMsgQuery, error) {
		id, err := uuid.Parse(v)
		if err != nil {
			return nil, err
		}
		return query.Where(entlog.HasInstanceWith(entinst.IDEQ(id))), nil
	},
	{
		field: "LEVEL",
		ftype: "MATCH",
	}: func(query *ent.LogMsgQuery, v string) (*ent.LogMsgQuery, error) {
		return query.Where(entlog.LevelEQ(v)), nil
	},
	{
		field: "LEVEL",
		ftype: "STARTING",
	}: func(query *ent.LogMsgQuery, v string) (*ent.LogMsgQuery, error) {
		levels := []string{"debug", "info", "error", "panic"}
		switch v {
		case "debug":
		case "info":
			levels = levels[1:]
		case "error":
			levels = levels[2:]
		case "panic":
			levels = levels[3:]
		}
		return query.Where(entlog.LevelIn(levels...)), nil
	},
}

func (flow *flow) ServerLogs(ctx context.Context, req *grpc.ServerLogsRequest) (*grpc.ServerLogsResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	clients := flow.edb.Clients(ctx)

	query := clients.LogMsg.Query()
	query = query.Where(entlog.Not(entlog.HasNamespace()), entlog.Not(entlog.HasWorkflow()))

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logEntFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.ServerLogsResponse)
	resp.PageInfo = pi

	resp.Results, err = bytedata.ConvertLogMsgForOutput(results)
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

	clients := flow.edb.Clients(ctx)
	query := clients.LogMsg.Query()
	query = query.Where(entlog.Not(entlog.HasNamespace()), entlog.Not(entlog.HasWorkflow()))

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logEntFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.ServerLogsResponse)
	resp.PageInfo = pi

	resp.Results, err = bytedata.ConvertLogMsgForOutput(results)
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

	cached := new(database.CacheData)

	err := flow.database.NamespaceByName(ctx, cached, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(ctx)

	query := clients.LogMsg.Query().Where(entlog.HasNamespaceWith(entns.ID(cached.Namespace.ID)))

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logEntFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.NamespaceLogsResponse)
	resp.Namespace = cached.Namespace.Name
	resp.PageInfo = pi

	resp.Results, err = bytedata.ConvertLogMsgForOutput(results)
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

	sub := flow.pubsub.SubscribeNamespaceLogs(&cached.Namespace.ID)
	defer flow.cleanup(sub.Close)

	clients := flow.edb.Clients(ctx)

resend:

	query := clients.LogMsg.Query().Where(entlog.HasNamespaceWith(entns.ID(cached.Namespace.ID)))

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logEntFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.NamespaceLogsResponse)
	resp.Namespace = cached.Namespace.Name
	resp.PageInfo = pi

	resp.Results, err = bytedata.ConvertLogMsgForOutput(results)
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

	cached := new(database.CacheData)

	err := flow.database.NamespaceByName(ctx, cached, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	err = flow.database.InodeByPath(ctx, cached, req.GetPath())
	if err != nil {
		return nil, err
	}

	err = flow.database.Workflow(ctx, cached, cached.Inode().Workflow)
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(ctx)

	query := clients.LogMsg.Query().Where(entlog.HasWorkflowWith(entwf.ID(cached.Workflow.ID)))

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logEntFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.WorkflowLogsResponse)
	resp.Namespace = cached.Namespace.Name
	resp.Path = cached.Path()
	resp.PageInfo = pi

	resp.Results, err = bytedata.ConvertLogMsgForOutput(results)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (flow *flow) WorkflowLogsParcels(req *grpc.WorkflowLogsRequest, srv grpc.Flow_WorkflowLogsParcelsServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	var tailing bool

	cached := new(database.CacheData)

	err := flow.database.NamespaceByName(ctx, cached, req.GetNamespace())
	if err != nil {
		return err
	}

	err = flow.database.InodeByPath(ctx, cached, req.GetPath())
	if err != nil {
		return err
	}

	err = flow.database.Workflow(ctx, cached, cached.Inode().Workflow)
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeWorkflowLogs(cached)
	defer flow.cleanup(sub.Close)

resend:

	clients := flow.edb.Clients(ctx)

	query := clients.LogMsg.Query().Where(entlog.HasWorkflowWith(entwf.ID(cached.Workflow.ID)))

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logEntFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.WorkflowLogsResponse)
	resp.Namespace = cached.Namespace.Name
	resp.Path = cached.Path()
	resp.PageInfo = pi

	resp.Results, err = bytedata.ConvertLogMsgForOutput(results)
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

	clients := flow.edb.Clients(ctx)
	// its important to append the intanceID to the callpath since we don't do it when creating the database entry
	prefix := internallogger.AppendInstanceID(cached.Instance.CallPath, cached.Instance.ID.String())
	root, err := internallogger.GetRootinstanceID(prefix)
	if err != nil {
		return nil, err
	}
	callerIsRoot := root == cached.Instance.Invoker

	query := clients.LogMsg.Query().Where(entlog.RootInstanceId(root))
	if !callerIsRoot {
		query = clients.LogMsg.Query().Where(entlog.And(entlog.RootInstanceIdEQ(root), entlog.LogInstanceCallPathHasPrefix(prefix)))
	}

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logEntFilters)
	if err != nil {
		return nil, err
	}

	filters := req.Pagination.Filter
	for _, v := range filters {
		results = queryJSON(v, results)
		pi.Total = int32(len(results))
	}

	resp := new(grpc.InstanceLogsResponse)
	resp.Namespace = cached.Namespace.Name
	resp.Instance = cached.Instance.ID.String()
	resp.PageInfo = pi
	resp.Results, err = bytedata.ConvertLogMsgForOutput(results)
	if err != nil {
		return nil, err
	}

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

	clients := flow.edb.Clients(ctx)
	query := clients.LogMsg.Query().Where(entlog.RootInstanceIdEQ(root))
	if !callerIsRoot {
		query = clients.LogMsg.Query().Where(entlog.And(entlog.RootInstanceIdEQ(root), entlog.LogInstanceCallPathHasPrefix(prefix)))
	}
	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logEntFilters)
	if err != nil {
		return err
	}

	filters := req.Pagination.Filter
	for _, v := range filters {
		results = queryJSON(v, results)
		pi.Total = int32(len(results))
	}

	resp := new(grpc.InstanceLogsResponse)
	resp.Namespace = cached.Namespace.Name
	resp.Instance = cached.Instance.ID.String()
	resp.PageInfo = pi

	resp.Results, err = bytedata.ConvertLogMsgForOutput(results)
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

func queryJSON(filter *grpc.PageFilter, results []*ent.LogMsg) []*ent.LogMsg {
	res := results
	if filter.Field == "QUERY" && filter.Type == "MATCH" {
		res = queryMatchState(filter.Val, results)
	}
	return res
}

func queryMatchState(q string, in []*ent.LogMsg) []*ent.LogMsg {
	values := strings.Split(q, "::")
	state := ""
	workflow := ""
	iterator := ""
	if len(values) >= 2 {
		workflow = values[0]
		state = values[1]
	}
	if len(values) > 2 {
		iterator = values[2]
	}
	res := make([]*ent.LogMsg, 0)
	for _, v := range in {
		if v.Tags["state-id"] == state &&
			v.Tags["workflow"] == workflow &&
			v.Tags["loop-index"] == iterator {
			res = append(res, v)
		}
	}
	return res
}
