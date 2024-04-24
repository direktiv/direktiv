package flow

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/direktiv/direktiv/pkg/refactor/mirror"
	"github.com/direktiv/direktiv/pkg/refactor/pubsub"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (flow *flow) CreateNamespaceMirror(ctx context.Context, req *grpc.CreateNamespaceMirrorRequest) (*grpc.CreateNamespaceResponse, error) {
	slog.Debug("Handling gRPC request", "this", this())

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

	ns, err = tx.DataStore().Namespaces().Create(ctx, &datastore.Namespace{
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
			slog.Error("pubsub publish", "error", err)
		}
	}()

	slog.Debug("Created namespace as git mirror", "namespace", ns.Name)

	var resp grpc.CreateNamespaceResponse
	resp.Namespace = bytedata.ConvertNamespaceToGrpc(ns)

	return &resp, nil
}

func (flow *flow) UpdateMirrorSettings(ctx context.Context, req *grpc.UpdateMirrorSettingsRequest) (*emptypb.Empty, error) {
	slog.Debug("Handling gRPC request", "this", this())

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

	slog.Debug("Updated mirror configs for namespace", "namespace", ns.Name)

	proc, err := flow.mirrorManager.NewProcess(ctx, ns, datastore.ProcessTypeSync)
	if err != nil {
		return nil, err
	}

	go func() {
		flow.mirrorManager.Execute(context.Background(), proc, mirConfig, &mirror.DirektivApplyer{NamespaceID: ns.ID})
		err := flow.pBus.Publish(pubsub.MirrorSync, ns.Name)
		if err != nil {
			slog.Error("pubsub publish", "error", err)
		}
	}()

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) SoftSyncMirror(ctx context.Context, req *grpc.SoftSyncMirrorRequest) (*emptypb.Empty, error) {
	slog.Debug("Handling gRPC request", "this", this())

	return flow.HardSyncMirror(ctx, &grpc.HardSyncMirrorRequest{
		Namespace: req.GetNamespace(),
		Path:      req.GetPath(),
	})
}

func (flow *flow) HardSyncMirror(ctx context.Context, req *grpc.HardSyncMirrorRequest) (*emptypb.Empty, error) {
	slog.Debug("Handling gRPC request", "this", this())

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
			slog.Error("pubsub publish", "error", err)
		}
	}()

	slog.Debug("Starting mirror process for namespace", "namespace", ns.Name)

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) MirrorInfo(ctx context.Context, req *grpc.MirrorInfoRequest) (*grpc.MirrorInfoResponse, error) {
	slog.Debug("Handling gRPC request", "this", this())

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
	slog.Debug("Handling gRPC request", "this", this())
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

func (flow *flow) CancelMirrorActivity(ctx context.Context, req *grpc.CancelMirrorActivityRequest) (*emptypb.Empty, error) {
	slog.Debug("Handling gRPC request", "this", this())

	mirProcessID, err := uuid.Parse(req.GetActivity())
	if err != nil {
		return nil, err
	}

	slog.Debug("cancelled by api request", "namespace", req.Namespace)
	flow.pubsub.CancelMirrorProcess(mirProcessID)

	// err = flow.mirrorManager.Cancel(ctx, mirProcessID)
	// if err != nil {
	// 	return nil, err
	// }
	var resp emptypb.Empty

	return &resp, nil
}
