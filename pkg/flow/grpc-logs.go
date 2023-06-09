package flow

import (
	"context"
	"strings"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/database/entwrapper"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	entlog "github.com/direktiv/direktiv/pkg/flow/ent/logmsg"
	entns "github.com/direktiv/direktiv/pkg/flow/ent/namespace"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/flow/internallogger"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/google/uuid"
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
		return query.Where(entlog.InstanceID(id)), nil
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

	query = query.Where(entlog.Not(entlog.HasNamespace()), entlog.WorkflowIDIsNil())

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logsFilters)
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
	query = query.Where(entlog.Not(entlog.HasNamespace()), entlog.Not(entlog.WorkflowID(uuid.UUID{})))

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logsFilters)
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

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logsFilters)
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

	sub := flow.pubsub.SubscribeNamespaceLogs(cached.Namespace.ID)
	defer flow.cleanup(sub.Close)

	clients := flow.edb.Clients(ctx)

resend:

	query := clients.LogMsg.Query().Where(entlog.HasNamespaceWith(entns.ID(cached.Namespace.ID)))

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logsFilters)
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

	ns, f, err := flow.getWorkflow(ctx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(ctx)

	query := clients.LogMsg.Query().Where(entlog.WorkflowID(f.ID))

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logsFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.WorkflowLogsResponse)
	resp.Namespace = ns.Name
	resp.Path = f.Path
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

	ns, f, err := flow.getWorkflow(ctx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeWorkflowLogs(f.ID)
	defer flow.cleanup(sub.Close)

resend:

	clients := flow.edb.Clients(ctx)

	query := clients.LogMsg.Query().Where(entlog.WorkflowID(f.ID))

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logsFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.WorkflowLogsResponse)
	resp.Namespace = ns.Name
	resp.Path = f.Path
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

	instance, err := flow.getInstance(ctx, req.GetNamespace(), req.GetInstance())
	if err != nil {
		return nil, err
	}

	// its important to append the instanceID to the callpath since we don't do it when creating the database entry
	prefix := internallogger.AppendInstanceID(instance.TelemetryInfo.CallPath, instance.Instance.ID.String())
	root, err := internallogger.GetRootinstanceID(prefix)
	if err != nil {
		return nil, err
	}
	callerIsRoot := root == instance.Instance.Invoker

	query := buildInstanceLogsQuery(ctx, flow.edb, root, prefix, callerIsRoot)
	logmsgs, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logsFilters)
	if err != nil {
		return nil, err
	}
	results, err := buildInstanceLogResp(ctx, logmsgs, pi, req.Pagination, req.Namespace, req.Instance)
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

	instance, err := flow.getInstance(ctx, req.GetNamespace(), req.GetInstance())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeInstanceLogs(instance.Instance.ID)
	defer flow.cleanup(sub.Close)

	// its important to append the intanceID to the callpath since we don't do it when creating the database entry.
	prefix := internallogger.AppendInstanceID(instance.TelemetryInfo.CallPath, instance.Instance.ID.String())
	root, err := internallogger.GetRootinstanceID(prefix)
	callerIsRoot := root == instance.Instance.ID.String()
	if err != nil {
		return err
	}

resend:

	query := buildInstanceLogsQuery(ctx, flow.edb, root, prefix, callerIsRoot)
	logmsgs, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logsFilters)
	if err != nil {
		return err
	}
	results, err := buildInstanceLogResp(ctx, logmsgs, pi, req.Pagination, req.Namespace, req.Instance)
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

		req.Pagination.Offset += int32(len(logmsgs))
	}

	more := sub.Wait(ctx)
	if !more {
		return nil
	}

	goto resend
}

// filters the passed *ent.LogMsg if the given filter is supported if
// the given filter is not supported returns the input unfiltered.
func filterLogmsg(filter *grpc.PageFilter, input []*ent.LogMsg) []*ent.LogMsg {
	res := input
	if filter.Field == "QUERY" && filter.Type == "MATCH" {
		res = filterMatchByWfStateIterator(filter.Val, input)
	}
	return res
}

// filters the input using the extracted values from the queryValue string.
// queryValue should be formatted like <workflow>::<state-id>::<loop-index>
// <state-id> and <indexId> is optional
// examples for queryValue:
// myworkflow or myworkflow:: or myworkflow::::
// myworkflow::getter or myworkflow::getter::
// myworkflow::getter::1
// ::getter::
// this method has two behaviors
// 1. if loop-index is left empty:
// when a logmsg from the input array has a matching pair of logtag values
// with the extracted values it will be added to the results
// 2: When the loop-index is provided:
// the result will contain all logmsg marked with the given
// loop-index starting the first match of the provided workflow and state-id
// additionally, all logmsgs from nested loops and childs will be added to the results.
func filterMatchByWfStateIterator(queryValue string, input []*ent.LogMsg) []*ent.LogMsg {
	values := strings.Split(queryValue, "::")
	state := ""
	workflow := ""
	iterator := ""
	if len(values) > 0 {
		workflow = values[0]
	}
	if len(values) > 1 {
		state = values[1]
	}
	if len(values) > 2 {
		iterator = values[2]
	}
	matchWf := make([]*ent.LogMsg, 0)
	matchState := make([]*ent.LogMsg, 0)
	matchIterator := make([]*ent.LogMsg, 0)
	for _, v := range input {
		if v.Tags["workflow"] == workflow {
			matchWf = append(matchWf, v)
		}
		if v.Tags["state-id"] == state &&
			workflow != "" && v.Tags["workflow"] == workflow {
			matchState = append(matchState, v)
		}
		if v.Tags["state-id"] == state &&
			workflow == "" {
			matchState = append(matchState, v)
		}
		if v.Tags["state-id"] != "" && v.Tags["state-id"] == state &&
			v.Tags["workflow"] == workflow &&
			v.Tags["loop-index"] == iterator {
			matchIterator = append(matchIterator, v)
		}
		if v.Tags["state-id"] == "" && v.Tags["workflow"] == workflow &&
			v.Tags["loop-index"] == iterator {
			matchIterator = append(matchIterator, v)
		}
	}
	if state == "" && iterator == "" {
		return matchWf
	}
	if workflow == "" && iterator == "" {
		return matchState
	}
	if iterator != "" {
		if len(matchIterator) == 0 {
			return make([]*ent.LogMsg, 0)
		}
		callpath := internallogger.AppendInstanceID(matchIterator[0].Tags["callpath"], matchIterator[0].Tags["instance-id"])
		childs := getAllChilds(callpath, input)
		originInstance := filterByInstanceId(matchIterator[0].Tags["instance-id"], input)
		subtree := append(originInstance, childs...)
		res := filterByIterrator(iterator, subtree)
		if nestedLoopHead := getNestedLoopHead(childs); nestedLoopHead != "" {
			nestedLoop := filterByInstanceId(nestedLoopHead, subtree)
			if len(nestedLoop) == 0 {
				return res
			}
			callpath := internallogger.AppendInstanceID(nestedLoop[0].Tags["callpath"], nestedLoop[0].Tags["instance-id"])
			nestedLoopChilds := getAllChilds(callpath, subtree)
			nestedLoopSubtree := append(nestedLoop, nestedLoopChilds...)
			res = append(res, nestedLoopChilds...)
			res = append(res, nestedLoopSubtree...)
			res = removeDuplicate(res)
		}
		return res
	}
	return matchState
}

func filterByIterrator(iterator string, in []*ent.LogMsg) []*ent.LogMsg {
	res := make([]*ent.LogMsg, 0)
	if iterator == "" {
		return res
	}
	for _, v := range in {
		if v.Tags["loop-index"] == iterator {
			res = append(res, v)
		}
	}
	return res
}

func getNestedLoopHead(in []*ent.LogMsg) string {
	for _, v := range in {
		if v.Tags["state-type"] == "foreach" {
			return v.Tags["instance-id"]
		}
	}
	return ""
}

func getAllChilds(callpath string, in []*ent.LogMsg) []*ent.LogMsg {
	res := make([]*ent.LogMsg, 0)
	for _, v := range in {
		if strings.HasPrefix(v.Tags["callpath"], callpath) {
			res = append(res, v)
		}
	}
	return res
}

func filterByInstanceId(instanceId string, in []*ent.LogMsg) []*ent.LogMsg {
	res := make([]*ent.LogMsg, 0)
	for _, v := range in {
		if strings.HasPrefix(v.Tags["instance-id"], instanceId) {
			res = append(res, v)
		}
	}
	return res
}

// https://stackoverflow.com/questions/66643946/how-to-remove-duplicates-strings-or-int-from-slice-in-go
func removeDuplicate(in []*ent.LogMsg) []*ent.LogMsg {
	allKeys := make(map[*ent.LogMsg]bool)
	list := []*ent.LogMsg{}
	for _, item := range in {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func buildInstanceLogResp(ctx context.Context,
	in []*ent.LogMsg,
	pi *grpc.PageInfo,
	page *grpc.Pagination,
	namespace string,
	instance string,
) (*grpc.InstanceLogsResponse, error) {
	filters := page.Filter
	results := in
	for _, v := range filters {
		results = filterLogmsg(v, in)
		pi.Total = int32(len(results))
	}

	resp := new(grpc.InstanceLogsResponse)
	resp.Namespace = namespace
	resp.Instance = instance
	resp.PageInfo = pi
	var err error
	resp.Results, err = bytedata.ConvertLogMsgForOutput(results)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func buildInstanceLogsQuery(ctx context.Context,
	edb *entwrapper.Database,
	root string,
	prefix string,
	callerIsRoot bool,
) *ent.LogMsgQuery {
	clients := edb.Clients(ctx)
	query := clients.LogMsg.Query().Where(entlog.RootInstanceIdEQ(root))
	if !callerIsRoot {
		query = clients.LogMsg.Query().Where(entlog.And(entlog.RootInstanceIdEQ(root), entlog.LogInstanceCallPathHasPrefix(prefix)))
	}
	return query
}
