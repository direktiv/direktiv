package flow

import (
	"context"
	"fmt"
	"strings"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/flow/internallogger"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/logengine"
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

const (
	namespaceType = "namespace"
	wf            = "workflow"
	ins           = "instance"
)

func addFiltersToQuery(query map[string]interface{}, filters ...*grpc.PageFilter) (map[string]interface{}, error) {
	for _, f := range filters {
		if f.Field == "ID" && f.Type == "MATCH" {
			id, err := uuid.Parse(f.Val)
			if err != nil {
				return nil, err
			}
			query["instance-id"] = id
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
	qu := make(map[string]interface{})
	qu["type"] = "server"
	qu, err := addFiltersToQuery(qu, req.Pagination.Filter...)
	if err != nil {
		return nil, err
	}
	total := 0
	err = flow.runSqlTx(ctx, func(tx *sqlTx) error {
		res, t, err := tx.DataStore().Logs().Get(ctx, qu, -1, -1)
		if err != nil {
			return err
		}
		total = t
		le = append(le, res...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	resp := new(grpc.ServerLogsResponse)
	resp.PageInfo = &grpc.PageInfo{Total: int32(total)}

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
	qu := make(map[string]interface{})
	qu["type"] = "server"
	qu, err := addFiltersToQuery(qu, req.Pagination.Filter...)
	if err != nil {
		return err
	}
	total := 0
	err = flow.runSqlTx(ctx, func(tx *sqlTx) error {
		res, t, err := tx.DataStore().Logs().Get(ctx, qu, int(req.Pagination.Limit), int(req.Pagination.Offset))
		if err != nil {
			return err
		}
		total = t
		le = append(le, res...)
		return nil
	})
	if err != nil {
		return err
	}
	resp := new(grpc.ServerLogsResponse)
	resp.PageInfo = &grpc.PageInfo{Limit: req.Pagination.Limit, Offset: req.Pagination.Offset, Total: int32(total)}
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

		req.Pagination.Offset += int32(len(le))
	}

	more := sub.Wait(ctx)
	if !more {
		return nil
	}

	goto resend
}

func (flow *flow) NamespaceLogs(ctx context.Context, req *grpc.NamespaceLogsRequest) (*grpc.NamespaceLogsResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	total := 0
	var err error
	var ns *core.Namespace
	le := make([]*logengine.LogEntry, 0)
	err = flow.runSqlTx(ctx, func(tx *sqlTx) error {
		ns, err = tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
		if err != nil {
			return err
		}

		qu := make(map[string]interface{})
		qu["source"] = ns.ID
		qu["type"] = namespaceType
		qu, err = addFiltersToQuery(qu, req.Pagination.Filter...)
		if err != nil {
			return err
		}

		res, t, err := tx.DataStore().Logs().Get(ctx, qu, int(req.Pagination.Limit), int(req.Pagination.Offset))
		if err != nil {
			return err
		}
		total = t
		le = append(le, res...)
		return nil
	})
	if err != nil {
		return nil, err
	}

	resp := new(grpc.NamespaceLogsResponse)
	resp.PageInfo = &grpc.PageInfo{Total: int32(total)}

	resp.Results, err = bytedata.ConvertLogMsgForOutput(le)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (flow *flow) NamespaceLogsParcels(req *grpc.NamespaceLogsRequest, srv grpc.Flow_NamespaceLogsParcelsServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	var err error
	var ns *core.Namespace
	err = flow.runSqlTx(ctx, func(tx *sqlTx) error {
		ns, err = tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
		return err
	})
	if err != nil {
		return err
	}

	var tailing bool

	sub := flow.pubsub.SubscribeNamespaceLogs(ns.ID)
	defer flow.cleanup(sub.Close)

resend:

	le := make([]*logengine.LogEntry, 0)
	qu := make(map[string]interface{})
	qu["source"] = ns.ID
	qu["type"] = namespaceType
	total := 0
	qu, err = addFiltersToQuery(qu, req.Pagination.Filter...)
	if err != nil {
		return err
	}
	err = flow.runSqlTx(ctx, func(tx *sqlTx) error {
		res, t, err := tx.DataStore().Logs().Get(ctx, qu, int(req.Pagination.Limit), int(req.Pagination.Offset))
		if err != nil {
			return err
		}
		total = t
		le = append(le, res...)
		return nil
	})
	if err != nil {
		return err
	}

	// leFiltered := logengine.FilterLogs(le, qu)
	resp := new(grpc.NamespaceLogsResponse)
	resp.PageInfo = &grpc.PageInfo{Limit: req.Pagination.Limit, Offset: req.Pagination.Offset, Total: int32(total)}
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
	qu["source"] = f.ID
	qu["type"] = wf
	qu, err = addFiltersToQuery(qu, req.Pagination.Filter...)
	if err != nil {
		return nil, err
	}
	le := make([]*logengine.LogEntry, 0)
	err = flow.runSqlTx(ctx, func(tx *sqlTx) error {
		res, _, err := tx.DataStore().Logs().Get(ctx, qu, int(req.Pagination.Limit), int(req.Pagination.Offset))
		if err != nil {
			return err
		}
		le = append(le, res...)
		return nil
	})
	if err != nil {
		return nil, err
	}
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
	qu["source"] = f.ID
	qu["type"] = wf
	qu, err = addFiltersToQuery(qu, req.Pagination.Filter...)
	if err != nil {
		return err
	}
	le := make([]*logengine.LogEntry, 0)
	err = flow.runSqlTx(ctx, func(tx *sqlTx) error {
		res, _, err := tx.DataStore().Logs().Get(ctx, qu, int(req.Pagination.Limit), int(req.Pagination.Offset))
		if err != nil {
			return err
		}
		le = append(le, res...)
		return nil
	})
	if err != nil {
		return err
	}

	resp := new(grpc.WorkflowLogsResponse)
	resp.Namespace = ns.Name
	resp.Path = f.Path
	resp.Results, err = bytedata.ConvertLogMsgForOutput(le)
	resp.PageInfo = &grpc.PageInfo{Total: int32(len(le))}
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

	instance, err := flow.getInstance(ctx, req.GetNamespace(), req.GetInstance())
	if err != nil {
		return nil, err
	}
	callpath := ""
	for _, v := range instance.DescentInfo.Descent {
		callpath += "/" + v.ID.String()
	}
	// its important to append the instanceID to the callpath since we don't do it when creating the database entry
	prefix := internallogger.AppendInstanceID(callpath, instance.Instance.ID.String())
	root, err := internallogger.GetRootinstanceID(prefix)
	if err != nil {
		return nil, err
	}
	qu := make(map[string]interface{})
	qu["log_instance_call_path"] = prefix
	qu["root_instance_id"] = root
	qu["type"] = ins
	qu, err = addFiltersToQuery(qu, req.Pagination.Filter...)
	if err != nil {
		return nil, err
	}
	le := make([]*logengine.LogEntry, 0)
	total := 0
	err = flow.runSqlTx(ctx, func(tx *sqlTx) error {
		res, t, err := tx.DataStore().Logs().Get(ctx, qu, int(req.Pagination.Limit), int(req.Pagination.Offset))
		if err != nil {
			return err
		}
		total = t
		le = append(le, res...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	leFiltered := filterLogs(le, qu)
	if _, ok := qu["loop-index"]; ok && len(leFiltered) > 0 {
		// special magic iterator stuff
		nestedLoop := false
		i := 0
		var e *logengine.LogEntry
		for i, e = range leFiltered {
			if e.Fields["state-type"] == "foreach" {
				nestedLoop = true
				break
			}
		}
		if nestedLoop {
			call := fmt.Sprintf("%v", leFiltered[i].Fields["callpath"])
			childs := getAllChilds(call, le)
			leFiltered = append(leFiltered, childs...)
		}
	}

	resp := new(grpc.InstanceLogsResponse)
	resp.Namespace = req.Namespace
	resp.Instance = req.Instance
	resp.PageInfo = &grpc.PageInfo{Total: int32(total)}
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

	instance, err := flow.getInstance(ctx, req.GetNamespace(), req.GetInstance())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeInstanceLogs(instance.Instance.ID)
	defer flow.cleanup(sub.Close)

	callpath := ""
	for _, v := range instance.DescentInfo.Descent {
		callpath += "/" + v.ID.String()
	}
	// its important to append the instanceID to the callpath since we don't do it when creating the database entry
	prefix := internallogger.AppendInstanceID(callpath, instance.Instance.ID.String())
	root, err := internallogger.GetRootinstanceID(prefix)
	if err != nil {
		return err
	}

resend:
	total := 0
	qu := make(map[string]interface{})
	qu["log_instance_call_path"] = prefix
	qu["root_instance_id"] = root
	qu["type"] = ins
	qu, err = addFiltersToQuery(qu, req.Pagination.Filter...)
	if err != nil {
		return err
	}
	le := make([]*logengine.LogEntry, 0)
	err = flow.runSqlTx(ctx, func(tx *sqlTx) error {
		res, t, err := tx.DataStore().Logs().Get(ctx, qu, int(req.Pagination.Limit), int(req.Pagination.Offset))
		if err != nil {
			return err
		}
		total = t
		le = append(le, res...)
		return nil
	})
	if err != nil {
		return err
	}
	leFiltered := filterLogs(le, qu)
	if _, ok := qu["loop-index"]; ok && len(leFiltered) > 0 {
		// special magic iterator stuff
		nestedLoop := false
		i := 0
		var e *logengine.LogEntry
		for i, e = range leFiltered {
			if e.Fields["state-type"] == "foreach" {
				nestedLoop = true
				break
			}
		}
		if nestedLoop {
			call := fmt.Sprintf("%v", leFiltered[i].Fields["callpath"])
			childs := getAllChilds(call, le)
			leFiltered = append(leFiltered, childs...)
		}
	}

	resp := new(grpc.InstanceLogsResponse)
	resp.Namespace = req.Namespace
	resp.Instance = req.Instance
	resp.PageInfo = &grpc.PageInfo{Total: int32(total)}
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

func filterLogs(logs []*logengine.LogEntry, keysAndValues map[string]interface{}) []*logengine.LogEntry {
	databaseCols := []string{
		"source",
		"log_instance_call_path",
		"type",
		"level",
		"root_instance_id",
	}

	for k := range keysAndValues { // the logstorer filters using db-cols
		for _, v2 := range databaseCols {
			if v2 == k {
				delete(keysAndValues, k)
			}
		}
	}
	filteredLogs := make([]*logengine.LogEntry, 0)

	for _, l := range logs {
		if shouldAdd(keysAndValues, l.Fields) {
			filteredLogs = append(filteredLogs, l)
		}
	}
	return filteredLogs
}

// returns true if all key values pairs are present in the fields and the values match.
// returns always true if keyAndValues is empty.
func shouldAdd(keysAndValues map[string]interface{}, fields map[string]interface{}) bool {
	match := true
	for k, e := range keysAndValues {
		t := fields[k]
		match = match && e == t
	}

	return match
}

func getAllChilds(callpath string, in []*logengine.LogEntry) []*logengine.LogEntry {
	res := make([]*logengine.LogEntry, 0)
	for _, v := range in {
		if strings.HasPrefix(fmt.Sprintf("%v", v.Fields["callpath"]), callpath) {
			res = append(res, v)
		}
	}
	return res
}

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
