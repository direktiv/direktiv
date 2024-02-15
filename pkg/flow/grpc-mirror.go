package flow

import (
	"context"
	"errors"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/direktiv/direktiv/pkg/refactor/logengine"
	"github.com/direktiv/direktiv/pkg/refactor/mirror"
	"github.com/direktiv/direktiv/pkg/refactor/pubsub"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (flow *flow) CreateNamespaceMirror(ctx context.Context, req *grpc.CreateNamespaceMirrorRequest) (*grpc.CreateNamespaceResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	settings := req.GetSettings()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetName())
	if err == nil && req.GetIdempotent() {
		var resp grpc.CreateNamespaceResponse
		resp.Namespace = bytedata.ConvertNamespaceToGrpc(ns)
		return &resp, nil
	}
	if !errors.Is(err, datastore.ErrNotFound) {
		return nil, err
	}

	ns, err = tx.DataStore().Namespaces().Create(ctx, &core.Namespace{
		Name: req.GetName(),
	})
	if err != nil {
		return nil, err
	}

	_, err = tx.FileStore().CreateRoot(ctx, uuid.New(), ns.Name)
	if err != nil {
		return nil, err
	}

	mirConfig, err := tx.DataStore().Mirror().CreateConfig(ctx, &datastore.MirrorConfig{
		Namespace:            ns.Name,
		GitRef:               settings.Ref,
		URL:                  settings.Url,
		PublicKey:            settings.PublicKey,
		PrivateKey:           settings.PrivateKey,
		PrivateKeyPassphrase: settings.Passphrase,
		Insecure:             settings.GetInsecure(),
	})
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	proc, err := flow.mirrorManager.NewProcess(ctx, ns, datastore.ProcessTypeInit)
	if err != nil {
		return nil, err
	}

	go func() {
		flow.mirrorManager.Execute(context.Background(), proc, mirConfig, &mirror.DirektivApplyer{NamespaceID: ns.ID})
		err := flow.pBus.Publish(pubsub.MirrorSync, ns.Name)
		if err != nil {
			flow.sugar.Error("pubsub publish", "error", err)
		}
	}()

	flow.logger.Infof(ctx, flow.ID, flow.GetAttributes(), "Created namespace as git mirror '%s'.", ns.Name)

	var resp grpc.CreateNamespaceResponse
	resp.Namespace = bytedata.ConvertNamespaceToGrpc(ns)

	return &resp, nil
}

func (flow *flow) CreateDirectoryMirror(ctx context.Context, req *grpc.CreateDirectoryMirrorRequest) (*grpc.CreateDirectoryResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	return nil, status.Error(codes.Unimplemented, "mirroring in directory is not allowed.")
}

func (flow *flow) UpdateMirrorSettings(ctx context.Context, req *grpc.UpdateMirrorSettingsRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	mirConfig, err := tx.DataStore().Mirror().GetConfig(ctx, ns.Name)
	if err != nil {
		return nil, err
	}

	settings := req.GetSettings()
	if s := settings.GetUrl(); s != "-" {
		mirConfig.URL = s
	}
	if s := settings.GetRef(); s != "-" {
		mirConfig.GitRef = s
	}
	if s := settings.GetPublicKey(); s != "-" {
		mirConfig.PublicKey = s
	}
	if s := settings.GetPrivateKey(); s != "-" {
		mirConfig.PrivateKey = s
	}
	if s := settings.GetPassphrase(); s != "-" {
		mirConfig.PrivateKeyPassphrase = s
	}
	if settings.Insecure != nil {
		mirConfig.Insecure = settings.GetInsecure()
	}

	mirConfig, err = tx.DataStore().Mirror().UpdateConfig(ctx, mirConfig)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	flow.logger.Infof(ctx, flow.ID, flow.GetAttributes(), "Updated mirror configs for namespace: %s", ns.Name)

	proc, err := flow.mirrorManager.NewProcess(ctx, ns, datastore.ProcessTypeSync)
	if err != nil {
		return nil, err
	}

	go func() {
		flow.mirrorManager.Execute(context.Background(), proc, mirConfig, &mirror.DirektivApplyer{NamespaceID: ns.ID})
		err := flow.pBus.Publish(pubsub.MirrorSync, ns.Name)
		if err != nil {
			flow.sugar.Error("pubsub publish", "error", err)
		}
	}()

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) LockMirror(ctx context.Context, req *grpc.LockMirrorRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	return nil, status.Error(codes.Unimplemented, "locking/unlocking mirror is not allowed.")
}

func (flow *flow) UnlockMirror(ctx context.Context, req *grpc.UnlockMirrorRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	return nil, status.Error(codes.Unimplemented, "locking/unlocking mirror is not allowed.")
}

func (flow *flow) SoftSyncMirror(ctx context.Context, req *grpc.SoftSyncMirrorRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	return flow.HardSyncMirror(ctx, &grpc.HardSyncMirrorRequest{
		Namespace: req.GetNamespace(),
		Path:      req.GetPath(),
	})
}

func (flow *flow) HardSyncMirror(ctx context.Context, req *grpc.HardSyncMirrorRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	mirConfig, err := tx.DataStore().Mirror().GetConfig(ctx, ns.Name)
	if err != nil {
		return nil, err
	}

	proc, err := flow.mirrorManager.NewProcess(ctx, ns, datastore.ProcessTypeSync)
	if err != nil {
		return nil, err
	}

	go func() {
		flow.mirrorManager.Execute(context.Background(), proc, mirConfig, &mirror.DirektivApplyer{NamespaceID: ns.ID})
		err := flow.pBus.Publish(pubsub.MirrorSync, ns.Name)
		if err != nil {
			flow.sugar.Error("pubsub publish", "error", err)
		}
	}()

	flow.logger.Infof(ctx, flow.ID, flow.GetAttributes(), "Starting mirror process for namespace: %s", ns.Name)

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) MirrorInfo(ctx context.Context, req *grpc.MirrorInfoRequest) (*grpc.MirrorInfoResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	mirConfig, err := tx.DataStore().Mirror().GetConfig(ctx, ns.Name)
	if err != nil {
		return nil, err
	}
	mirProcesses, err := tx.DataStore().Mirror().GetProcessesByNamespace(ctx, ns.Name)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.MirrorInfoResponse)
	resp.Namespace = ns.Name
	resp.Info = bytedata.ConvertMirrorConfigToGrpcMirrorInfo(mirConfig)
	resp.Activities = new(grpc.MirrorActivities)
	resp.Activities.PageInfo = nil
	resp.Activities.Results = bytedata.ConvertMirrorProcessesToGrpcMirrorActivityInfoList(mirProcesses)

	if mirConfig.PrivateKeyPassphrase != "" {
		resp.Info.Passphrase = "-"
	}
	if mirConfig.PrivateKey != "" {
		resp.Info.PrivateKey = "-"
	}

	return resp, nil
}

func (flow *flow) MirrorInfoStream(req *grpc.MirrorInfoRequest, srv grpc.Flow_MirrorInfoStreamServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())
	ctx := srv.Context()

	resp, err := flow.MirrorInfo(ctx, req)
	if err != nil {
		return err
	}
	// mock streaming response.
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			err = srv.Send(resp)
			if err != nil {
				return err
			}
			time.Sleep(time.Second * 5)
		}
	}
}

func (flow *flow) MirrorActivityLogs(ctx context.Context, req *grpc.MirrorActivityLogsRequest) (*grpc.MirrorActivityLogsResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	mirProcess, err := tx.DataStore().Mirror().GetProcess(ctx, ns.ID)
	if err != nil {
		return nil, err
	}

	qu := make(map[string]interface{})
	qu["source"] = mirProcess
	qu, err = addFiltersToQuery(qu, req.Pagination.Filter...)
	if err != nil {
		return nil, err
	}
	le := make([]*logengine.LogEntry, 0)
	res, total, err := tx.DataStore().Logs().Get(ctx, qu, int(req.Pagination.Limit), int(req.Pagination.Offset))
	if err != nil {
		return nil, err
	}
	le = append(le, res...)

	resp := new(grpc.MirrorActivityLogsResponse)
	resp.Namespace = ns.Name
	resp.Activity = mirProcess.ID.String()
	resp.PageInfo = &grpc.PageInfo{Limit: req.Pagination.Limit, Offset: req.Pagination.Offset, Total: int32(total)}
	resp.Results, err = bytedata.ConvertLogMsgForOutput(le)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (flow *flow) MirrorActivityLogsParcels(req *grpc.MirrorActivityLogsRequest, srv grpc.Flow_MirrorActivityLogsParcelsServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	mirProcessID, err := uuid.Parse(req.GetActivity())
	if err != nil {
		return err
	}

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return err
	}

	var tailing bool

	mirProcess, err := tx.DataStore().Mirror().GetProcess(ctx, mirProcessID)
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeMirrorActivityLogs(ns.ID, mirProcess.ID)
	defer flow.cleanup(sub.Close)

resend:
	qu := make(map[string]interface{})
	qu["source"] = mirProcessID
	qu, err = addFiltersToQuery(qu, req.Pagination.Filter...)
	if err != nil {
		return err
	}
	le := make([]*logengine.LogEntry, 0)
	res, total, err := tx.DataStore().Logs().Get(ctx, qu, int(req.Pagination.Limit), int(req.Pagination.Offset))
	if err != nil {
		return err
	}
	le = append(le, res...)

	if err != nil {
		return err
	}

	resp := new(grpc.MirrorActivityLogsResponse)
	resp.Namespace = ns.Name
	resp.Activity = mirProcessID.String()
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

		req.Pagination.Offset += int32(len(resp.Results))
	}

	more := sub.Wait(ctx)
	if !more {
		return nil
	}

	goto resend
}

func (flow *flow) CancelMirrorActivity(ctx context.Context, req *grpc.CancelMirrorActivityRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	mirProcessID, err := uuid.Parse(req.GetActivity())
	if err != nil {
		return nil, err
	}

	flow.logger.Debugf(ctx, flow.ID, flow.GetAttributes(), "cancelled by api request")
	flow.pubsub.CancelMirrorProcess(mirProcessID)

	// err = flow.mirrorManager.Cancel(ctx, mirProcessID)
	// if err != nil {
	// 	return nil, err
	// }
	var resp emptypb.Empty

	return &resp, nil
}
