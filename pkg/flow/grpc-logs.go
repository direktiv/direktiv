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

	// TODO: Add special magic iterator query
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

	// TODO: Add special magic iterator query
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
}
