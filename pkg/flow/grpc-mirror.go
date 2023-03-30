package flow

// TODO: yassir, need refactor.
/*

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	entlog "github.com/direktiv/direktiv/pkg/flow/ent/logmsg"
	entmir "github.com/direktiv/direktiv/pkg/flow/ent/mirror"
	entmiract "github.com/direktiv/direktiv/pkg/flow/ent/mirroractivity"
)

func (flow *flow) CreateNamespaceMirror(ctx context.Context, req *grpc.CreateNamespaceMirrorRequest) (*grpc.CreateNamespaceResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	var ino *ent.Inode
	var mir *ent.Mirror
	var ns *ent.Namespace

	clients := flow.edb.Clients(tctx)

	settings := req.GetSettings()

	if req.GetIdempotent() {

		cached := new(database.CacheData)

		err = flow.database.NamespaceByName(tctx, cached, req.GetName())
		if err != nil {
			rollback(tx)
			goto respond
		}
		if !derrors.IsNotFound(err) {
			return nil, err
		}
	}

	ns, err = clients.Namespace.Create().SetName(req.GetName()).Save(tctx)
	if err != nil {
		return nil, err
	}

	ino, err = clients.Inode.Create().SetNillableName(nil).SetType(util.InodeTypeDirectory).SetExtendedType(util.InodeTypeGit).SetReadOnly(true).SetNamespaceID(ns.ID).Save(tctx)
	if err != nil {
		return nil, err
	}

	mir, err = clients.Mirror.Create().
		SetURL(settings.GetUrl()).
		SetRef(settings.GetRef()).
		SetCommit("").
		SetPublicKey(settings.GetPublicKey()).
		SetPrivateKey(settings.GetPrivateKey()).
		SetPassphrase(settings.GetPassphrase()).
		SetCron(settings.GetCron()).
		SetNillableLastSync(nil).
		SetInode(ino).
		SetNamespaceID(ns.ID).
		Save(tctx)
	if err != nil {
		return nil, err
	}

	err = flow.syncer.NewActivity(tx, &newMirrorActivityArgs{
		MirrorID: mir.ID.String(),
		Type:     util.MirrorActivityTypeInit,
	})
	if err != nil {
		flow.logger.Errorf(ctx, flow.ID, flow.GetAttributes(), "failed to create git mirror %v", err)
		return nil, err
	}
	flow.logger.Infof(ctx, flow.ID, flow.GetAttributes(), "Created namespace as git mirror '%s'.", ns.Name)
	flow.pubsub.NotifyNamespaces()

respond:

	var resp grpc.CreateNamespaceResponse

	err = bytedata.ConvertDataForOutput(ns, &resp.Namespace)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (flow *flow) CreateDirectoryMirror(ctx context.Context, req *grpc.CreateDirectoryMirrorRequest) (*grpc.CreateDirectoryResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	namespace := req.GetNamespace()

	path := GetInodePath(req.GetPath())
	dir, base := filepath.Split(path)

	if base == "" || base == "/" {
		return nil, status.Error(codes.AlreadyExists, "root directory already exists")
	}

	cached, err := flow.traverseToInode(tctx, namespace, dir)
	if err != nil {
		return nil, err
	}

	if cached.Inode().Type != util.InodeTypeDirectory {
		return nil, status.Error(codes.AlreadyExists, "parent node is not a directory")
	}

	if cached.Inode().ReadOnly {
		return nil, errors.New("cannot write into read-only directory")
	}

	settings := req.GetSettings()
	var mir *ent.Mirror

	clients := flow.edb.Clients(tctx)

	ino, err := clients.Inode.Create().SetName(base).SetNamespaceID(cached.Namespace.ID).SetParentID(cached.Inode().ID).SetType(util.InodeTypeDirectory).SetExtendedType(util.InodeTypeGit).SetReadOnly(true).Save(tctx)
	if err != nil {
		return nil, err
	}

	mir, err = clients.Mirror.Create().
		SetURL(settings.GetUrl()).
		SetRef(settings.GetRef()).
		SetCommit("").
		SetPublicKey(settings.GetPublicKey()).
		SetPrivateKey(settings.GetPrivateKey()).
		SetPassphrase(settings.GetPassphrase()).
		SetCron(settings.GetCron()).
		SetNillableLastSync(nil).
		SetInode(ino).
		SetNamespaceID(cached.Namespace.ID).
		Save(tctx)
	if err != nil {
		return nil, err
	}

	err = flow.syncer.NewActivity(tx, &newMirrorActivityArgs{
		MirrorID: mir.ID.String(),
		Type:     util.MirrorActivityTypeInit,
	})
	if err != nil {
		flow.logger.Errorf(ctx, flow.ID, flow.GetAttributes(), "Failed to create directory as git mirror %s", path)
		return nil, err
	}

	flow.logger.Infof(ctx, cached.Namespace.ID, cached.GetAttributes(recipient.Namespace), "Created directory as git mirror '%s'.", path)
	flow.pubsub.NotifyInode(cached.Inode())

	var resp grpc.CreateDirectoryResponse

	err = bytedata.ConvertDataForOutput(ino, &resp.Node)
	if err != nil {
		return nil, err
	}

	resp.Node.ReadOnly = true

	resp.Namespace = namespace
	resp.Node.Parent = dir
	resp.Node.Path = path

	// Broadcast
	err = flow.BroadcastDirectory(ctx, BroadcastEventTypeCreate,
		broadcastDirectoryInput{
			Path:   resp.Node.Path,
			Parent: resp.Node.Parent,
		}, cached)

	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (srv *server) getMirror(ctx context.Context, ino *database.Inode) (*database.Mirror, error) {
	if ino.ExtendedType != util.InodeTypeGit {
		srv.sugar.Debugf("%s inode isn't a git mirror", parent())
		return nil, ErrNotMirror
	}

	mir, err := srv.database.Mirror(ctx, ino.Mirror)
	if err != nil {
		srv.sugar.Debugf("%s failed to query inode's mirror: %v", parent(), err)
		return nil, err
	}

	return mir, nil
}

func (srv *server) traverseToMirror(ctx context.Context, namespace, path string) (*database.CacheData, *database.Mirror, error) {
	cached, err := srv.traverseToInode(ctx, namespace, path)
	if err != nil {
		srv.sugar.Debugf("%s failed to resolve mirror's inode: %v", parent(), err)
		return nil, nil, err
	}

	mir, err := srv.getMirror(ctx, cached.Inode())
	if err != nil {
		return nil, nil, err
	}

	return cached, mir, nil
}

func (flow *flow) UpdateMirrorSettings(ctx context.Context, req *grpc.UpdateMirrorSettingsRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, mirror, err := flow.traverseToMirror(tctx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	settings := req.GetSettings()

	clients := flow.edb.Clients(tctx)

	updater := clients.Mirror.UpdateOneID(mirror.ID)

	if s := settings.GetUrl(); s != "-" {
		updater.SetURL(s)
	}

	if s := settings.GetRef(); s != "-" {
		updater.SetRef(s)
	}

	if s := settings.GetPublicKey(); s != "-" {
		updater.SetPublicKey(s)
	}

	if s := settings.GetPrivateKey(); s != "-" {
		updater.SetPrivateKey(s)
	}

	if s := settings.GetPassphrase(); s != "-" {
		updater.SetPassphrase(s)
	}

	if s := settings.GetCron(); s != "-" {
		updater.SetCron(s)
	}

	x, err := updater.Save(ctx)
	if err != nil {
		return nil, err
	}

	err = flow.syncer.NewActivity(tx, &newMirrorActivityArgs{
		MirrorID: x.ID.String(),
		Type:     util.MirrorActivityTypeReconfigure,
	})
	if err != nil {
		return nil, err
	}

	flow.pubsub.NotifyMirror(cached.Inode())

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) LockMirror(ctx context.Context, req *grpc.LockMirrorRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, mirror, err := flow.traverseToMirror(tctx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	if !cached.Inode().ReadOnly {
		return nil, ErrMirrorLocked
	}

	ino := cached.Inode()
	updatedInodes := make([]*database.Inode, 0)

	var recurser func(ino *database.Inode) error
	recurser = func(ino *database.Inode) error {
		for _, child := range ino.Children {
			if ino.ExtendedType == util.InodeTypeGit {
				continue
			}

			readonly := false

			ino, err = flow.database.UpdateInode(tctx, &database.UpdateInodeArgs{
				Inode:    child,
				ReadOnly: &readonly,
			})
			if err != nil {
				return err
			}

			updatedInodes = append(updatedInodes, ino)

			if ino.Type == util.InodeTypeDirectory {
				err = recurser(ino)
				if err != nil {
					return err
				}
			} else if ino.Type == util.InodeTypeWorkflow {

				cached := new(database.CacheData)

				err = flow.database.Inode(tctx, cached, ino.ID)
				if err != nil {
					return err
				}

				err = flow.database.Workflow(tctx, cached, cached.Inode().Workflow)
				if err != nil {
					return err
				}

				readonly := false

				_, err = flow.database.UpdateWorkflow(tctx, &database.UpdateWorkflowArgs{
					ID:       cached.Workflow.ID,
					ReadOnly: &readonly,
				})
				if err != nil {
					return err
				}

			} else {
				return errors.New("inode type unaccounted for")
			}
		}

		return nil
	}

	readonly := false

	ino, err = flow.database.UpdateInode(tctx, &database.UpdateInodeArgs{
		Inode:    ino,
		ReadOnly: &readonly,
	})
	if err != nil {
		return nil, err
	}

	updatedInodes = append(updatedInodes, ino)

	err = recurser(ino)
	if err != nil {
		return nil, err
	}

	err = flow.syncer.NewActivity(tx, &newMirrorActivityArgs{
		MirrorID: mirror.ID.String(),
		Type:     util.MirrorActivityTypeLocked,
	})
	if err != nil {
		return nil, err
	}

	flow.pubsub.NotifyMirror(cached.Inode())
	for _, uino := range updatedInodes {
		flow.pubsub.NotifyInode(uino)
	}

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) UnlockMirror(ctx context.Context, req *grpc.UnlockMirrorRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, mirror, err := flow.traverseToMirror(tctx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	if cached.Inode().ReadOnly {
		return nil, ErrMirrorUnlocked
	}

	ino := cached.Inode()
	updatedInodes := make([]*database.Inode, 0)

	var recurser func(ino *database.Inode) error
	recurser = func(ino *database.Inode) error {
		for _, child := range ino.Children {
			if ino.ExtendedType == util.InodeTypeGit {
				continue
			}

			readonly := false

			ino, err := flow.database.UpdateInode(tctx, &database.UpdateInodeArgs{
				Inode:    child,
				ReadOnly: &readonly,
			})
			if err != nil {
				return err
			}

			updatedInodes = append(updatedInodes, ino)

			if ino.Type == util.InodeTypeDirectory {
				err = recurser(ino)
				if err != nil {
					return err
				}
			} else if ino.Type == util.InodeTypeWorkflow {

				cached := new(database.CacheData)

				err = flow.database.Inode(tctx, cached, ino.ID)
				if err != nil {
					return err
				}

				err = flow.database.Workflow(tctx, cached, cached.Inode().Workflow)
				if err != nil {
					return err
				}

				readonly := false

				_, err = flow.database.UpdateWorkflow(tctx, &database.UpdateWorkflowArgs{
					ID:       cached.Workflow.ID,
					ReadOnly: &readonly,
				})
				if err != nil {
					return err
				}

			} else {
				return errors.New("inode type unaccounted for")
			}
		}

		return nil
	}

	readonly := false

	x, err := flow.database.UpdateInode(tctx, &database.UpdateInodeArgs{
		Inode:    ino,
		ReadOnly: &readonly,
	})
	if err != nil {
		return nil, err
	}

	updatedInodes = append(updatedInodes, x)

	err = recurser(ino)
	if err != nil {
		return nil, err
	}

	err = flow.syncer.NewActivity(tx, &newMirrorActivityArgs{
		MirrorID: mirror.ID.String(),
		Type:     util.MirrorActivityTypeUnlocked,
	})
	if err != nil {
		return nil, err
	}

	flow.pubsub.NotifyMirror(cached.Inode())
	for _, uino := range updatedInodes {
		flow.pubsub.NotifyInode(uino)
	}

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) SoftSyncMirror(ctx context.Context, req *grpc.SoftSyncMirrorRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, mirror, err := flow.traverseToMirror(tctx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	err = flow.syncer.NewActivity(tx, &newMirrorActivityArgs{
		MirrorID: mirror.ID.String(),
		Type:     util.MirrorActivityTypeSync,
	})
	if err != nil {
		return nil, err
	}

	flow.pubsub.NotifyMirror(cached.Inode())

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) HardSyncMirror(ctx context.Context, req *grpc.HardSyncMirrorRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, mirror, err := flow.traverseToMirror(tctx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	err = flow.syncer.NewActivity(tx, &newMirrorActivityArgs{
		MirrorID: mirror.ID.String(),
		Type:     util.MirrorActivityTypeSync,
	})
	if err != nil {
		return nil, err
	}

	flow.pubsub.NotifyMirror(cached.Inode())

	var resp emptypb.Empty

	return &resp, nil
}

var mirrorActivitiesOrderings = []*orderingInfo{
	{
		db:           entmiract.FieldCreatedAt,
		req:          "CREATED",
		defaultOrder: ent.Asc,
	},
}

var mirrorActivitiesFilters = map[*filteringInfo]func(query *ent.MirrorActivityQuery, v string) (*ent.MirrorActivityQuery, error){}

func (flow *flow) MirrorInfo(ctx context.Context, req *grpc.MirrorInfoRequest) (*grpc.MirrorInfoResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, mirror, err := flow.traverseToMirror(tctx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(tctx)

	query := clients.MirrorActivity.Query().Where(entmiract.HasMirrorWith(entmir.ID(mirror.ID)))

	results, pi, err := paginate[*ent.MirrorActivityQuery, *ent.MirrorActivity](ctx, req.Pagination, query, mirrorActivitiesOrderings, mirrorActivitiesFilters)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	resp := new(grpc.MirrorInfoResponse)
	resp.Namespace = cached.Namespace.Name
	resp.Activities = new(grpc.MirrorActivities)
	resp.Activities.PageInfo = pi

	err = bytedata.ConvertDataForOutput(results, &resp.Activities.Results)
	if err != nil {
		return nil, err
	}

	err = bytedata.ConvertDataForOutput(mirror, &resp.Info)
	if err != nil {
		return nil, err
	}

	if mirror.Passphrase != "" {
		resp.Info.Passphrase = "-"
	}
	if mirror.PrivateKey != "" {
		resp.Info.PrivateKey = "-"
	}

	return resp, nil
}

func (flow *flow) MirrorInfoStream(req *grpc.MirrorInfoRequest, srv grpc.Flow_MirrorInfoStreamServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	cached, _, err := flow.traverseToMirror(ctx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return err
	}

	if cached.Inode().ExtendedType != util.InodeTypeGit {
		return ErrNotMirror
	}

	sub := flow.pubsub.SubscribeMirror(cached)
	defer flow.cleanup(sub.Close)

	var mirror *database.Mirror

resend:

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx)

	cached, mirror, err = flow.traverseToMirror(tctx, req.GetNamespace(), cached.Path())
	if err != nil {
		return err
	}

	clients := flow.edb.Clients(tctx)

	query := clients.MirrorActivity.Query().Where(entmiract.HasMirrorWith(entmir.ID(mirror.ID)))

	results, pi, err := paginate[*ent.MirrorActivityQuery, *ent.MirrorActivity](tctx, req.Pagination, query, mirrorActivitiesOrderings, mirrorActivitiesFilters)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	resp := new(grpc.MirrorInfoResponse)
	resp.Namespace = cached.Namespace.Name
	resp.Activities = new(grpc.MirrorActivities)
	resp.Activities.PageInfo = pi

	err = bytedata.ConvertDataForOutput(results, &resp.Activities.Results)
	if err != nil {
		return err
	}

	err = bytedata.ConvertDataForOutput(mirror, &resp.Info)
	if err != nil {
		return err
	}

	if mirror.Passphrase != "" {
		resp.Info.Passphrase = "-"
	}
	if mirror.PrivateKey != "" {
		resp.Info.PrivateKey = "-"
	}

	nhash = bytedata.Checksum(resp)
	if nhash != phash {
		err = srv.Send(resp)
		if err != nil {
			return err
		}
	}
	phash = nhash

	more := sub.Wait(ctx)
	if !more {
		return nil
	}

	goto resend
}

func (srv *server) getMirrorActivity(ctx context.Context, namespace, activity string) (*database.CacheData, *database.MirrorActivity, error) {
	id, err := uuid.Parse(activity)
	if err != nil {
		srv.sugar.Debugf("%s failed to parse UUID: %v", parent(), err)
		return nil, nil, err
	}

	cached := new(database.CacheData)

	err = srv.database.NamespaceByName(ctx, cached, namespace)
	if err != nil {
		srv.sugar.Debugf("%s failed to resolve namespace: %v", parent(), err)
		return nil, nil, err
	}

	act, err := srv.database.MirrorActivity(ctx, id)
	if err != nil {
		srv.sugar.Debugf("%s failed to query instance: %v", parent(), err)
		return nil, nil, err
	}

	if act.Namespace != cached.Namespace.ID {
		return nil, nil, os.ErrNotExist
	}

	return cached, act, nil
}

func (flow *flow) MirrorActivityLogs(ctx context.Context, req *grpc.MirrorActivityLogsRequest) (*grpc.MirrorActivityLogsResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached, activity, err := flow.getMirrorActivity(ctx, req.GetNamespace(), req.GetActivity())
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(ctx)

	query := clients.LogMsg.Query().Where(entlog.HasActivityWith(entmiract.ID(activity.ID)))

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logsFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.MirrorActivityLogsResponse)
	resp.Namespace = cached.Namespace.Name
	resp.Activity = activity.ID.String()
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

	var tailing bool

	cached, activity, err := flow.getMirrorActivity(ctx, req.GetNamespace(), req.GetActivity())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeMirrorActivityLogs(cached.Namespace, activity)
	defer flow.cleanup(sub.Close)

resend:

	clients := flow.edb.Clients(ctx)

	query := clients.LogMsg.Query().Where(entlog.HasActivityWith(entmiract.ID(activity.ID)))

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logsFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.MirrorActivityLogsResponse)
	resp.Namespace = cached.Namespace.Name
	resp.Activity = activity.ID.String()
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
	flow.logger.Debugf(ctx, flow.ID, flow.GetAttributes(), "cancelled by api request")
	flow.syncer.cancelActivity(req.GetActivity(), "cancel.api", "cancelled by api request")

	var resp emptypb.Empty

	return &resp, nil
}
*/
