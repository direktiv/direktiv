package flow

import (
	"context"
	"strings"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/flow/internallogger"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/logengine"
	"github.com/google/uuid"
)

const (
	ns string = "namespace"
	wf string = "workflow"
)

type fileAttributes filestore.File

func (f fileAttributes) GetAttributes() map[string]interface{} {
	m := make(map[string]interface{})
	m["namespace_logs"] = f.RootID.String()
	if f.Typ == filestore.FileTypeWorkflow {
		m["workflow_id"] = f.ID.String()
	}
	return m
}

func addFiltersToQuery(query map[string]interface{}, filters ...*grpc.PageFilter) (map[string]interface{}, error) {
	for _, f := range filters {
		if f.Field == "ID" && f.Type == "MATCH" {
			id, err := uuid.Parse(f.Val)
			if err != nil {
				return nil, err
			}
			query["instance_logs"] = id
		}
		if f.Field == "LEVEL" && f.Type == "STARTING" {
			level := logengine.Debug
			switch f.Val {
			case "debug":
				level = logengine.Debug
			case "info":
				level = logengine.Info
			case "error":
				level = logengine.Error
			}
			query["level"] = level
		}
		if f.Field == "QUERY" && f.Type == "MATCH" {
			values := strings.Split(f.Val, "::")
			if len(values) > 0 && values[0] != "" {
				query["workflow"] = values[0]
			}
			if len(values) > 1 && values[1] != "" {
				query["state-id"] = values[1]
			}
			if len(values) > 2 && values[2] != "" {
				query["loop-index"] = values[2]
			}
		}
	}
	return query, nil
}

func (flow *flow) ServerLogs(ctx context.Context, req *grpc.ServerLogsRequest) (*grpc.ServerLogsResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())
	le := make([]*logengine.LogEntry, 0)
	err := flow.runSqlTx(ctx, func(fStore filestore.FileStore, store datastore.Store) error {
		qu := make(map[string]interface{})
		qu["sender_type"] = "server"
		qu, err := addFiltersToQuery(qu, req.Pagination.Filter...)
		if err != nil {
			return err
		}
		res, err := store.Logs().Get(ctx, qu, -1, -1)
		if err != nil {
			return err
		}
		le = append(le, res...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	resp := new(grpc.ServerLogsResponse)
	resp.PageInfo = &grpc.PageInfo{Total: int32(len(le))}

	resp.Results, err = bytedata.ConvertLogMsgForOutput(le)
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

	le := make([]*logengine.LogEntry, 0)
	err := flow.runSqlTx(ctx, func(fStore filestore.FileStore, store datastore.Store) error {
		qu := make(map[string]interface{})
		qu["sender_type"] = "server"
		qu, err := addFiltersToQuery(qu, req.Pagination.Filter...)
		if err != nil {
			return err
		}
		res, err := store.Logs().Get(ctx, qu, int(req.Pagination.Limit), int(req.Pagination.Offset))
		if err != nil {
			return err
		}
		le = append(le, res...)
		return nil
	})
	if err != nil {
		return err
	}
	resp := new(grpc.ServerLogsResponse)
	resp.PageInfo = &grpc.PageInfo{Limit: req.Pagination.Limit, Offset: req.Pagination.Offset, Total: int32(len(le))}
	resp.Results, err = bytedata.ConvertLogMsgForOutput(le)
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

	le := make([]*logengine.LogEntry, 0)
	qu := make(map[string]interface{})
	qu["namespace_logs"] = cached.Namespace.ID
	qu["sender_type"] = ns
	qu, err = addFiltersToQuery(qu, req.Pagination.Filter...)
	if err != nil {
		return nil, err
	}
	err = flow.runSqlTx(ctx, func(fStore filestore.FileStore, store datastore.Store) error {
		res, err := store.Logs().Get(ctx, qu, -1, -1)
		if err != nil {
			return err
		}
		le = append(le, res...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	le = logengine.FilterLogs(le, qu)
	resp := new(grpc.NamespaceLogsResponse)
	resp.PageInfo = &grpc.PageInfo{Total: int32(len(le))}

	resp.Results, err = bytedata.ConvertLogMsgForOutput(le)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (flow *flow) NamespaceLogsParcels(req *grpc.NamespaceLogsRequest, srv grpc.Flow_NamespaceLogsParcelsServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	cached := new(database.CacheData)

	err := flow.database.NamespaceByName(ctx, cached, req.GetNamespace())
	if err != nil {
		return err
	}

	var tailing bool

	sub := flow.pubsub.SubscribeNamespaceLogs(&cached.Namespace.ID)
	defer flow.cleanup(sub.Close)

resend:

	le := make([]*logengine.LogEntry, 0)
	qu := make(map[string]interface{})
	qu["namespace_logs"] = cached.Namespace.ID
	qu["sender_type"] = ns
	qu, err = addFiltersToQuery(qu, req.Pagination.Filter...)
	err = flow.runSqlTx(ctx, func(fStore filestore.FileStore, store datastore.Store) error {
		if err != nil {
			return err
		}
		res, err := store.Logs().Get(ctx, qu, -1, -1)
		if err != nil {
			return err
		}
		le = append(le, res...)
		return nil
	})

	if err != nil {
		return err
	}
	leFiltered := logengine.FilterLogs(le, qu)
	resp := new(grpc.NamespaceLogsResponse)
	resp.PageInfo = &grpc.PageInfo{Limit: req.Pagination.Limit, Offset: req.Pagination.Offset, Total: int32(len(leFiltered))}
	resp.Results, err = bytedata.ConvertLogMsgForOutput(leFiltered)
	if err != nil {
		return err
	}

	if len(resp.Results) != 0 || !tailing {
		tailing = true

		err = srv.Send(resp)
		if err != nil {
			return err
		}

		req.Pagination.Offset += int32(len(le))
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
	qu := make(map[string]interface{})
	qu["workflow_id"] = f.ID
	qu["sender_type"] = wf
	qu, err = addFiltersToQuery(qu, req.Pagination.Filter...)
	if err != nil {
		return nil, err
	}
	le := make([]*logengine.LogEntry, 0)
	err = flow.runSqlTx(ctx, func(fStore filestore.FileStore, store datastore.Store) error {
		res, err := store.Logs().Get(ctx, qu, -1, -1)
		if err != nil {
			return err
		}
		le = append(le, res...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	le = logengine.FilterLogs(le, qu)
	resp := new(grpc.WorkflowLogsResponse)
	resp.Namespace = ns.Name
	resp.Path = f.Path
	resp.PageInfo = &grpc.PageInfo{Total: int32(len(le))}
	resp.Results, err = bytedata.ConvertLogMsgForOutput(le)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (flow *flow) WorkflowLogsParcels(req *grpc.WorkflowLogsRequest, srv grpc.Flow_WorkflowLogsParcelsServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	ns, f, err := flow.getWorkflow(ctx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeWorkflowLogs(f.ID)
	defer flow.cleanup(sub.Close)

	var tailing bool

resend:
	qu := make(map[string]interface{})
	qu["workflow_id"] = f.ID
	qu["sender_type"] = wf
	qu, err = addFiltersToQuery(qu, req.Pagination.Filter...)
	if err != nil {
		return err
	}
	le := make([]*logengine.LogEntry, 0)
	err = flow.runSqlTx(ctx, func(fStore filestore.FileStore, store datastore.Store) error {
		res, err := store.Logs().Get(ctx, qu, int(req.Pagination.Limit), int(req.Pagination.Offset))
		if err != nil {
			return err
		}
		le = append(le, res...)
		return nil
	})
	if err != nil {
		return err
	}
	leFiltered := logengine.FilterLogs(le, qu)

	resp := new(grpc.WorkflowLogsResponse)
	resp.Namespace = ns.Name
	resp.Path = f.Path
	resp.Results, err = bytedata.ConvertLogMsgForOutput(leFiltered)
	if err != nil {
		return err
	}

	if len(resp.Results) != 0 || !tailing {
		tailing = true

		err = srv.Send(resp)
		if err != nil {
			return err
		}

		req.Pagination.Offset += int32(len(le))
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
	prefix := internallogger.AppendInstanceID(cached.Instance.CallPath, cached.Instance.ID.String())
	root, err := internallogger.GetRootinstanceID(prefix)
	if err != nil {
		return nil, err
	}
	qu := make(map[string]interface{})
	qu["log_instance_call_path"] = prefix
	qu["root_instance_id"] = root
	qu, err = addFiltersToQuery(qu, req.Pagination.Filter...)
	if err != nil {
		return nil, err
	}
	le := make([]*logengine.LogEntry, 0)
	err = flow.runSqlTx(ctx, func(fStore filestore.FileStore, store datastore.Store) error {
		res, err := store.Logs().Get(ctx, qu, -1, -1)
		if err != nil {
			return err
		}
		le = append(le, res...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	leFiltered := logengine.FilterLogs(le, qu)
	resp := new(grpc.InstanceLogsResponse)
	resp.Namespace = cached.Namespace.Name
	resp.Instance = cached.Instance.ID.String()
	resp.PageInfo = &grpc.PageInfo{Total: int32(len(leFiltered))}
	resp.Results, err = bytedata.ConvertLogMsgForOutput(leFiltered)
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
	prefix := internallogger.AppendInstanceID(cached.Instance.CallPath, cached.Instance.ID.String())
	root, err := internallogger.GetRootinstanceID(prefix)
	if err != nil {
		return err
	}
resend:
	qu := make(map[string]interface{})
	qu["log_instance_call_path"] = prefix
	qu["root_instance_id"] = root
	qu, err = addFiltersToQuery(qu, req.Pagination.Filter...)
	if err != nil {
		return err
	}
	le := make([]*logengine.LogEntry, 0)
	err = flow.runSqlTx(ctx, func(fStore filestore.FileStore, store datastore.Store) error {
		res, err := store.Logs().Get(ctx, qu, int(req.Pagination.Limit), int(req.Pagination.Offset))
		if err != nil {
			return err
		}
		le = append(le, res...)
		return nil
	})
	if err != nil {
		return err
	}
	leFiltered := logengine.FilterLogs(le, qu)
	resp := new(grpc.InstanceLogsResponse)
	resp.Namespace = cached.Namespace.Name
	resp.Instance = cached.Instance.ID.String()
	resp.PageInfo = &grpc.PageInfo{Limit: req.Pagination.Limit, Offset: req.Pagination.Offset, Total: int32(len(leFiltered))}

	resp.Results, err = bytedata.ConvertLogMsgForOutput(leFiltered)
	if err != nil {
		return err
	}

	if len(resp.Results) != 0 || !tailing {
		tailing = true

		err = srv.Send(resp)
		if err != nil {
			return err
		}

		req.Pagination.Offset += int32(len(le))
	}

	more := sub.Wait(ctx)
	if !more {
		return nil
	}

	goto resend

	// 	// its important to append the intanceID to the callpath since we don't do it when creating the database entry.
	// 	root, err := internallogger.GetRootinstanceID(prefix)
	// 	callerIsRoot := root == cached.Instance.ID.String()
	// 	if err != nil {
	// 		return err
	// 	}

	// resend:

	// 	query := buildInstanceLogsQuery(ctx, flow.edb, root, prefix, callerIsRoot)
	// 	logmsgs, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logsFilters)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	results, err := buildInstanceLogResp(ctx, logmsgs, pi, req.Pagination, req.Namespace, req.Instance)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	resp := results

	// 	if len(resp.Results) != 0 || !tailing {
	// 		tailing = true

	// 		err = srv.Send(resp)
	// 		if err != nil {
	// 			return err
	// 		}

	// 		req.Pagination.Offset += int32(len(logmsgs))
	// 	}

	// 	more := sub.Wait(ctx)
	// 	if !more {
	// 		return nil
	// 	}
}

// filters the passed *ent.LogMsg if the given filter is supported if
// the given filter is not supported returns the input unfiltered.
// func filterLogmsg(filter *grpc.PageFilter, input []*ent.LogMsg) []*ent.LogMsg {
// 	res := input
// 	if filter.Field == "QUERY" && filter.Type == "MATCH" {
// 		res = filterMatchByWfStateIterator(filter.Val, input)
// 	}
// 	return res
// }

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
// func filterMatchByWfStateIterator(queryValue string, input []*logengine.LogEntry) []*logengine.LogEntry {
// 	values := strings.Split(queryValue, "::")
// 	state := ""
// 	workflow := ""
// 	iterator := ""
// 	if len(values) > 0 {
// 		workflow = values[0]
// 	}
// 	if len(values) > 1 {
// 		state = values[1]
// 	}
// 	if len(values) > 2 {
// 		iterator = values[2]
// 	}
// 	matchWf := make([]*logengine.LogEntry, 0)
// 	matchState := make([]*logengine.LogEntry, 0)
// 	matchIterator := make([]*logengine.LogEntry, 0)
// 	for _, v := range input {
// 		if v.Fields["workflow"] == workflow {
// 			matchWf = append(matchWf, v)
// 		}
// 		if v.Fields["state-id"] == state &&
// 			workflow != "" && v.Fields["workflow"] == workflow {
// 			matchState = append(matchState, v)
// 		}
// 		if v.Fields["state-id"] == state &&
// 			workflow == "" {
// 			matchState = append(matchState, v)
// 		}
// 		if v.Fields["state-id"] != "" && v.Fields["state-id"] == state &&
// 			v.Fields["workflow"] == workflow &&
// 			v.Fields["loop-index"] == iterator {
// 			matchIterator = append(matchIterator, v)
// 		}
// 		if v.Fields["state-id"] == "" && v.Fields["workflow"] == workflow &&
// 			v.Fields["loop-index"] == iterator {
// 			matchIterator = append(matchIterator, v)
// 		}
// 	}
// 	if state == "" && iterator == "" {
// 		return matchWf
// 	}
// 	if workflow == "" && iterator == "" {
// 		return matchState
// 	}
// 	if iterator != "" {
// 		if len(matchIterator) == 0 {
// 			return make([]*logengine.LogEntry, 0)
// 		}
// 		callpath := internallogger.AppendInstanceID(matchIterator[0].Fields["callpath"], matchIterator[0].Fields["instance-id"])
// 		childs := getAllChilds(callpath, input)
// 		originInstance := filterByInstanceId(matchIterator[0].Fields["instance-id"], input)
// 		subtree := append(originInstance, childs...)
// 		res := filterByIterrator(iterator, subtree)
// 		if nestedLoopHead := getNestedLoopHead(childs); nestedLoopHead != "" {
// 			nestedLoop := filterByInstanceId(nestedLoopHead, subtree)
// 			if len(nestedLoop) == 0 {
// 				return res
// 			}
// 			callpath := internallogger.AppendInstanceID(nestedLoop[0].Tags["callpath"], nestedLoop[0].Tags["instance-id"])
// 			nestedLoopChilds := getAllChilds(callpath, subtree)
// 			nestedLoopSubtree := append(nestedLoop, nestedLoopChilds...)
// 			res = append(res, nestedLoopChilds...)
// 			res = append(res, nestedLoopSubtree...)
// 			res = removeDuplicate(res)
// 		}
// 		return res
// 	}
// 	return matchState
// }

// func filterByIterrator(iterator string, in []*logengine.LogEntry) []*logengine.LogEntry {
// 	res := make([]*logengine.LogEntry, 0)
// 	if iterator == "" {
// 		return res
// 	}
// 	for _, v := range in {
// 		if v.Fields["loop-index"] == iterator {
// 			res = append(res, v)
// 		}
// 	}
// 	return res
// }

// func getNestedLoopHead(in []*logengine.LogEntry) string {
// 	for _, v := range in {
// 		if v.Fields["state-type"] == "foreach" {
// 			return v.Fields["instance-id"]
// 		}
// 	}
// 	return ""
// }

// func getAllChilds(callpath string, in []*logengine.LogEntry) []*logengine.LogEntry {
// 	res := make([]*logengine.LogEntry, 0)
// 	for _, v := range in {
// 		if strings.HasPrefix(v.Fields["callpath"], callpath) {
// 			res = append(res, v)
// 		}
// 	}
// 	return res
// }

// func filterByInstanceId(instanceId string, in []*logengine.LogEntry) []*logengine.LogEntry {
// 	res := make([]*logengine.LogEntry, 0)
// 	for _, v := range in {
// 		if strings.HasPrefix(v.Fields["instance-id"], instanceId) {
// 			res = append(res, v)
// 		}
// 	}
// 	return res
// }

// https://stackoverflow.com/questions/66643946/how-to-remove-duplicates-strings-or-int-from-slice-in-go
// func removeDuplicate(in []*logengine.LogEntry) []*logengine.LogEntry {
// 	allKeys := make(map[*logengine.LogEntry]bool)
// 	list := []*logengine.LogEntry{}
// 	for _, item := range in {
// 		if _, value := allKeys[item]; !value {
// 			allKeys[item] = true
// 			list = append(list, item)
// 		}
// 	}
// 	return list
// }

// func buildInstanceLogResp(ctx context.Context,
// 	in []*ent.LogMsg,
// 	pi *grpc.PageInfo,
// 	page *grpc.Pagination,
// 	namespace string,
// 	instance string,
// ) (*grpc.InstanceLogsResponse, error) {
// 	filters := page.Filter
// 	results := in
// 	for _, v := range filters {
// 		results = filterLogmsg(v, in)
// 		pi.Total = int32(len(results))
// 	}

// 	resp := new(grpc.InstanceLogsResponse)
// 	resp.Namespace = namespace
// 	resp.Instance = instance
// 	resp.PageInfo = pi
// 	var err error
// 	resp.Results, err = bytedata.ConvertLogMsgForOutput(results)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return resp, nil
// }
