package flow

import (
	"context"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	entlog "github.com/direktiv/direktiv/pkg/flow/ent/logmsg"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/mirror"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (flow *flow) CreateNamespaceMirror(ctx context.Context, req *grpc.CreateNamespaceMirrorRequest) (*grpc.CreateNamespaceResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx, tx, err := flow.edb.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	var ns *ent.Namespace

	clients := flow.edb.Clients(ctx)

	settings := req.GetSettings()

	if req.GetIdempotent() {
		ns, err := flow.edb.NamespaceByName(ctx, req.GetName())
		if err == nil {
			var resp grpc.CreateNamespaceResponse
			err = bytedata.ConvertDataForOutput(ns, &resp.Namespace)
			if err != nil {
				return nil, err
			}

			return &resp, nil
		}
		if !derrors.IsNotFound(err) {
			return nil, err
		}
	}

	ns, err = clients.Namespace.Create().SetName(req.GetName()).Save(ctx)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	// create namespace filesystem root and mirror config.
	var txErr error
	var mirConfig *mirror.Config
	err = flow.runSqlTx(ctx, func(fStore filestore.FileStore, store datastore.Store) error {
		var root *filestore.Root
		root, txErr = fStore.CreateRoot(ctx, ns.ID)
		if txErr != nil {
			return txErr
		}
		_, _, txErr = fStore.ForRootID(root.ID).CreateFile(ctx, "/", filestore.FileTypeDirectory, nil)
		if txErr != nil {
			return txErr
		}

		mirConfig, txErr = store.Mirror().CreateConfig(ctx, &mirror.Config{
			NamespaceID:          ns.ID,
			GitRef:               settings.Ref,
			URL:                  settings.Url,
			PublicKey:            settings.PublicKey,
			PrivateKey:           settings.PrivateKey,
			PrivateKeyPassphrase: settings.Passphrase,
		})
		if txErr != nil {
			return txErr
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	_, err = flow.mirrorManager.StartInitialMirroringProcess(ctx, mirConfig)
	if err != nil {
		return nil, err
	}
	flow.logger.Infof(ctx, flow.ID, flow.GetAttributes(), "Created namespace as git mirror '%s'.", ns.Name)

	var resp grpc.CreateNamespaceResponse
	err = bytedata.ConvertDataForOutput(ns, &resp.Namespace)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (flow *flow) CreateDirectoryMirror(ctx context.Context, req *grpc.CreateDirectoryMirrorRequest) (*grpc.CreateDirectoryResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	return nil, status.Error(codes.Unimplemented, "mirroring in directory is not allowed.")
}

func (flow *flow) UpdateMirrorSettings(ctx context.Context, req *grpc.UpdateMirrorSettingsRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	ctx, tx, err := flow.edb.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	_, store, commit, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	mirConfig, err := store.Mirror().GetConfig(ctx, ns.ID)
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

	mirConfig, err = store.Mirror().UpdateConfig(ctx, mirConfig)
	if err != nil {
		return nil, err
	}

	if err = commit(ctx); err != nil {
		return nil, err
	}

	flow.logger.Infof(ctx, flow.ID, flow.GetAttributes(), "Updated mirror configs for namespace: %s", ns.Name)

	_, err = flow.mirrorManager.StartSyncingMirrorProcess(ctx, mirConfig)
	if err != nil {
		return nil, err
	}

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

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	_, store, _, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	mirConfig, err := store.Mirror().GetConfig(ctx, ns.ID)
	if err != nil {
		return nil, err
	}

	_, err = flow.mirrorManager.StartSyncingMirrorProcess(ctx, mirConfig)
	if err != nil {
		return nil, err
	}

	flow.logger.Infof(ctx, flow.ID, flow.GetAttributes(), "Starting mirror process for namespace: %s", ns.Name)

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) MirrorInfo(ctx context.Context, req *grpc.MirrorInfoRequest) (*grpc.MirrorInfoResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	_, store, _, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	mirConfig, err := store.Mirror().GetConfig(ctx, ns.ID)
	if err != nil {
		return nil, err
	}
	mirProcesses, err := store.Mirror().GetProcessesByNamespaceID(ctx, mirConfig.NamespaceID)
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

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	_, store, _, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	mirProcess, err := store.Mirror().GetProcess(ctx, ns.ID)
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(ctx)

	query := clients.LogMsg.Query().Where(entlog.MirrorActivityID(mirProcess.ID))

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logsFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.MirrorActivityLogsResponse)
	resp.Namespace = ns.Name
	resp.Activity = mirProcess.ID.String()
	resp.PageInfo = pi

	err = bytedata.ConvertDataForOutput(results, &resp.Results)
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

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return err
	}

	var tailing bool

	_, store, _, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return err
	}
	defer rollback()

	mirProcess, err := store.Mirror().GetProcess(ctx, mirProcessID)
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeMirrorActivityLogs(ns.ID, mirProcess.ID)
	defer flow.cleanup(sub.Close)

resend:

	clients := flow.edb.Clients(ctx)

	query := clients.LogMsg.Query().Where(entlog.MirrorActivityID(mirProcess.ID))

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logsFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.MirrorActivityLogsResponse)
	resp.Namespace = ns.Name
	resp.Activity = mirProcess.ID.String()
	resp.PageInfo = pi

	err = bytedata.ConvertDataForOutput(results, &resp.Results)
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
	err = flow.mirrorManager.CancelMirroringProcess(ctx, mirProcessID)
	if err != nil {
		return nil, err
	}
	var resp emptypb.Empty

	return &resp, nil
}
