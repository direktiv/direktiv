package flow

import (
	"context"
	"fmt"
	"strings"

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

// filters the passed *internallogger.LogMsgs if the given filter is supported if
// the given filter is not supported returns the input unfiltered.
func filterLogmsg(filter *grpc.PageFilter, input []*internallogger.LogMsgs) []*internallogger.LogMsgs {
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
func filterMatchByWfStateIterator(queryValue string, input []*internallogger.LogMsgs) []*internallogger.LogMsgs {
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
	matchWf := make([]*internallogger.LogMsgs, 0)
	matchState := make([]*internallogger.LogMsgs, 0)
	matchIterator := make([]*internallogger.LogMsgs, 0)
	for _, v := range input {
		if v.Tags["workflow"] == nil {
			v.Tags["workflow"] = ""
		}
		if v.Tags["state-id"] == nil {
			v.Tags["state-id"] = ""
		}
		if v.Tags["loop-index"] == nil {
			v.Tags["loop-index"] = ""
		}
		vWorkflow := fmt.Sprintf("%s", v.Tags["workflow"])
		vStateid := fmt.Sprintf("%s", v.Tags["state-id"])
		vLoopindex := fmt.Sprintf("%s", v.Tags["loop-index"])
		if v.Tags["workflow"] == workflow {
			matchWf = append(matchWf, v)
		}
		if vStateid == state &&
			workflow != "" && vWorkflow == workflow {
			matchState = append(matchState, v)
		}
		if vStateid == state &&
			workflow == "" {
			matchState = append(matchState, v)
		}
		if vStateid != "" && vStateid == state &&
			vWorkflow == workflow &&
			vLoopindex == iterator {
			matchIterator = append(matchIterator, v)
		}
		if vStateid == "" && vWorkflow == workflow &&
			vLoopindex == iterator {
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
			return make([]*internallogger.LogMsgs, 0)
		}
		if matchIterator[0].Tags["callpath"] == nil {
			matchIterator[0].Tags["callpath"] = ""
		}
		vcallpath := fmt.Sprintf("%v", matchIterator[0].Tags["callpath"])
		if matchIterator[0].Tags["instance-id"] == nil {
			matchIterator[0].Tags["instance-id"] = ""
		}
		vInstanceid := fmt.Sprintf("%v", matchIterator[0].Tags["instance-id"])
		callpath := internallogger.AppendInstanceID(vcallpath, vInstanceid)
		childs := getAllChilds(callpath, input)
		originInstance := filterByInstanceId(fmt.Sprintf("%v", matchIterator[0].Tags["instance-id"]), input)
		subtree := append(originInstance, childs...)
		res := filterByIterrator(iterator, subtree)
		if nestedLoopHead := getNestedLoopHead(childs); nestedLoopHead != "" {
			nestedLoop := filterByInstanceId(nestedLoopHead, subtree)
			if len(nestedLoop) == 0 {
				return res
			}
			if nestedLoop[0].Tags["callpath"] == nil {
				nestedLoop[0].Tags["callpath"] = ""
			}
			if nestedLoop[0].Tags["instance-id"] == nil {
				nestedLoop[0].Tags["instance-id"] = ""
			}
			r := fmt.Sprintf("%s", nestedLoop[0].Tags["callpath"])
			c := fmt.Sprintf("%s", nestedLoop[0].Tags["instance-id"])
			callpath := internallogger.AppendInstanceID(r, c)
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

func filterByIterrator(iterator string, in []*internallogger.LogMsgs) []*internallogger.LogMsgs {
	res := make([]*internallogger.LogMsgs, 0)
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

func getNestedLoopHead(in []*internallogger.LogMsgs) string {
	for _, v := range in {
		if v.Tags["state-type"] == "foreach" {
			return fmt.Sprintf("%v", v.Tags["instance-id"])
		}
	}
	return ""
}

func getAllChilds(callpath string, in []*internallogger.LogMsgs) []*internallogger.LogMsgs {
	res := make([]*internallogger.LogMsgs, 0)
	for _, v := range in {
		if strings.HasPrefix(fmt.Sprintf("%v", v.Tags["callpath"]), callpath) {
			res = append(res, v)
		}
	}
	return res
}

func filterByInstanceId(instanceId string, in []*internallogger.LogMsgs) []*internallogger.LogMsgs {
	res := make([]*internallogger.LogMsgs, 0)
	for _, v := range in {
		if v.Tags["instance-id"] == instanceId {
			res = append(res, v)
		}
	}
	return res
}

// https://stackoverflow.com/questions/66643946/how-to-remove-duplicates-strings-or-int-from-slice-in-go
func removeDuplicate(in []*internallogger.LogMsgs) []*internallogger.LogMsgs {
	allKeys := make(map[*internallogger.LogMsgs]bool)
	list := []*internallogger.LogMsgs{}
	for _, item := range in {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
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
		results = filterLogmsg(v, in)
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
