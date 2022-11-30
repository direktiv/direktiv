package flow

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/ent"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	entmiract "github.com/direktiv/direktiv/pkg/flow/ent/mirroractivity"
)

func (flow *flow) CreateNamespaceMirror(ctx context.Context, req *grpc.CreateNamespaceMirrorRequest) (*grpc.CreateNamespaceResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace
	inoc := tx.Inode
	mirc := tx.Mirror
	var ns *ent.Namespace
	var ino *ent.Inode
	var mir *ent.Mirror

	settings := req.GetSettings()

	if req.GetIdempotent() {
		ns, err = flow.getNamespace(ctx, nsc, req.GetName())
		if err == nil {
			rollback(tx)
			goto respond
		}
		if !derrors.IsNotFound(err) {
			return nil, err
		}
	}

	ns, err = nsc.Create().SetName(req.GetName()).Save(ctx)
	if err != nil {
		return nil, err
	}

	ino, err = inoc.Create().SetNillableName(nil).SetType(util.InodeTypeDirectory).SetExtendedType(util.InodeTypeGit).SetReadOnly(true).SetNamespace(ns).Save(ctx)
	if err != nil {
		return nil, err
	}

	mir, err = mirc.Create().
		SetURL(settings.GetUrl()).
		SetRef(settings.GetRef()).
		SetCommit("").
		SetPublicKey(settings.GetPublicKey()).
		SetPrivateKey(settings.GetPrivateKey()).
		SetPassphrase(settings.GetPassphrase()).
		SetCron(settings.GetCron()).
		SetNillableLastSync(nil).
		SetInode(ino).
		SetNamespace(ns).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	err = flow.syncer.NewActivity(tx, &newMirrorActivityArgs{
		MirrorID: mir.ID.String(),
		Type:     util.MirrorActivityTypeInit,
	})
	if err != nil {
		return nil, err
	}

	flow.logToServer(ctx, time.Now(), "Created namespace as git mirror '%s'.", ns.Name)
	flow.pubsub.NotifyNamespaces()

respond:

	var resp grpc.CreateNamespaceResponse

	err = atob(ns, &resp.Namespace)
	if err != nil {
		return nil, err
	}

	return &resp, nil

}

func (flow *flow) CreateDirectoryMirror(ctx context.Context, req *grpc.CreateDirectoryMirrorRequest) (*grpc.CreateDirectoryResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	namespace := req.GetNamespace()
	ns, err := flow.getNamespace(ctx, tx.Namespace, namespace)
	if err != nil {
		return nil, err
	}

	path := GetInodePath(req.GetPath())
	dir, base := filepath.Split(path)

	if base == "" || base == "/" {
		return nil, status.Error(codes.AlreadyExists, "root directory already exists")
	}

	inoc := tx.Inode

	pino, err := flow.getInode(ctx, inoc, ns, dir, req.GetParents())
	if err != nil {
		return nil, err
	}

	if pino.ino.Type != util.InodeTypeDirectory {
		return nil, status.Error(codes.AlreadyExists, "parent node is not a directory")
	}

	if pino.ino.ReadOnly {
		return nil, errors.New("cannot write into read-only directory")
	}

	settings := req.GetSettings()
	mirc := tx.Mirror
	var mir *ent.Mirror

	ino, err := inoc.Create().SetName(base).SetNamespace(ns).SetParent(pino.ino).SetType(util.InodeTypeDirectory).SetExtendedType(util.InodeTypeGit).SetReadOnly(true).Save(ctx)
	if err != nil {
		return nil, err
	}

	mir, err = mirc.Create().
		SetURL(settings.GetUrl()).
		SetRef(settings.GetRef()).
		SetCommit("").
		SetPublicKey(settings.GetPublicKey()).
		SetPrivateKey(settings.GetPrivateKey()).
		SetPassphrase(settings.GetPassphrase()).
		SetCron(settings.GetCron()).
		SetNillableLastSync(nil).
		SetInode(ino).
		SetNamespace(ns).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	err = flow.syncer.NewActivity(tx, &newMirrorActivityArgs{
		MirrorID: mir.ID.String(),
		Type:     util.MirrorActivityTypeInit,
	})
	if err != nil {
		return nil, err
	}

	flow.logToNamespace(ctx, time.Now(), ns, "Created directory as git mirror '%s'.", path)
	flow.pubsub.NotifyInode(pino.ino)

	// respond:

	var resp grpc.CreateDirectoryResponse

	err = atob(ino, &resp.Node)
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
		}, ns)

	if err != nil {
		return nil, err
	}

	return &resp, nil

}

func (srv *server) getMirror(ctx context.Context, ino *ent.Inode) (*ent.Mirror, error) {

	if ino.ExtendedType != util.InodeTypeGit {
		srv.sugar.Debugf("%s inode isn't a git mirror", parent())
		return nil, ErrNotMirror
	}

	mir, err := ino.QueryMirror().Only(ctx)
	if err != nil {
		srv.sugar.Debugf("%s failed to query inode's mirror: %v", parent(), err)
		return nil, err
	}

	return mir, nil

}

type mirData struct {
	*nodeData
	mir *ent.Mirror
}

func (srv *server) traverseToMirror(ctx context.Context, nsc *ent.NamespaceClient, namespace, path string) (*mirData, error) {

	nd, err := srv.traverseToInode(ctx, nsc, namespace, path)
	if err != nil {
		srv.sugar.Debugf("%s failed to resolve mirror's inode: %v", parent(), err)
		return nil, err
	}

	md := new(mirData)
	md.nodeData = nd

	mir, err := srv.getMirror(ctx, md.ino)
	if err != nil {
		srv.sugar.Debugf("%s failed to get mirror: %v", parent(), err)
		return nil, err
	}

	md.mir = mir

	md.ino.Edges.Namespace = md.ns()
	// NOTE: can't do this due to cycle: wd.ino.Edges.Workflow = wf
	mir.Edges.Inode = md.ino
	mir.Edges.Namespace = md.ns()

	return md, nil

}

func (flow *flow) UpdateMirrorSettings(ctx context.Context, req *grpc.UpdateMirrorSettingsRequest) (*emptypb.Empty, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	d, err := flow.traverseToMirror(ctx, tx.Namespace, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	settings := req.GetSettings()

	updater := d.mir.Update()

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

	edges := d.mir.Edges

	d.mir, err = updater.Save(ctx)
	if err != nil {
		return nil, err
	}

	d.mir.Edges = edges

	err = flow.syncer.NewActivity(tx, &newMirrorActivityArgs{
		MirrorID: d.mir.ID.String(),
		Type:     util.MirrorActivityTypeReconfigure,
	})
	if err != nil {
		return nil, err
	}

	// flow.logToNamespace(ctx, time.Now(), ns, "Created directory as git mirror '%s'.", path)
	flow.pubsub.NotifyMirror(d.ino)

	// respond:

	var resp emptypb.Empty

	return &resp, nil

}

func (flow *flow) LockMirror(ctx context.Context, req *grpc.LockMirrorRequest) (*emptypb.Empty, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	d, err := flow.traverseToMirror(ctx, tx.Namespace, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	if !d.ino.ReadOnly {
		return nil, ErrMirrorLocked
	}

	ino := d.ino
	updatedInodes := make([]*ent.Inode, 0)

	var recurser func(ino *ent.Inode) error
	recurser = func(ino *ent.Inode) error {

		inos, err := ino.QueryChildren().All(ctx)
		if err != nil {
			return err
		}

		for _, ino := range inos {
			if ino.ExtendedType == util.InodeTypeGit {
				continue
			}

			ino, err := ino.Update().SetReadOnly(false).Save(ctx)
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
				wf, err := ino.QueryWorkflow().Only(ctx)
				if err != nil {
					return err
				}
				_, err = wf.Update().SetReadOnly(false).Save(ctx)
				if err != nil {
					return err
				}
			} else {
				return errors.New("inode type unaccounted for")
			}
		}

		return nil

	}

	ino, err = ino.Update().SetReadOnly(false).Save(ctx)
	if err != nil {
		return nil, err
	}

	updatedInodes = append(updatedInodes, ino)

	err = recurser(ino)
	if err != nil {
		return nil, err
	}

	err = flow.syncer.NewActivity(tx, &newMirrorActivityArgs{
		MirrorID: d.mir.ID.String(),
		Type:     util.MirrorActivityTypeLocked,
	})
	if err != nil {
		return nil, err
	}

	flow.pubsub.NotifyMirror(d.ino)
	for _, uino := range updatedInodes {
		flow.pubsub.NotifyInode(uino)
	}

	var resp emptypb.Empty

	return &resp, nil

}

func (flow *flow) UnlockMirror(ctx context.Context, req *grpc.UnlockMirrorRequest) (*emptypb.Empty, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	d, err := flow.traverseToMirror(ctx, tx.Namespace, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	if d.ino.ReadOnly {
		return nil, ErrMirrorUnlocked
	}

	ino := d.ino
	updatedInodes := make([]*ent.Inode, 0)

	var recurser func(ino *ent.Inode) error
	recurser = func(ino *ent.Inode) error {

		inos, err := ino.QueryChildren().All(ctx)
		if err != nil {
			return err
		}

		for _, ino := range inos {
			if ino.ExtendedType == util.InodeTypeGit {
				continue
			}

			ino, err := ino.Update().SetReadOnly(true).Save(ctx)
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
				wf, err := ino.QueryWorkflow().Only(ctx)
				if err != nil {
					return err
				}
				_, err = wf.Update().SetReadOnly(true).Save(ctx)
				if err != nil {
					return err
				}
			} else {
				return errors.New("inode type unaccounted for")
			}
		}

		return nil

	}

	ino, err = ino.Update().SetReadOnly(true).Save(ctx)
	if err != nil {
		return nil, err
	}

	updatedInodes = append(updatedInodes, ino)

	err = recurser(ino)
	if err != nil {
		return nil, err
	}

	err = flow.syncer.NewActivity(tx, &newMirrorActivityArgs{
		MirrorID: d.mir.ID.String(),
		Type:     util.MirrorActivityTypeUnlocked,
	})
	if err != nil {
		return nil, err
	}

	flow.pubsub.NotifyMirror(d.ino)
	for _, uino := range updatedInodes {
		flow.pubsub.NotifyInode(uino)
	}

	var resp emptypb.Empty

	return &resp, nil

}

func (flow *flow) SoftSyncMirror(ctx context.Context, req *grpc.SoftSyncMirrorRequest) (*emptypb.Empty, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	d, err := flow.traverseToMirror(ctx, tx.Namespace, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	err = flow.syncer.NewActivity(tx, &newMirrorActivityArgs{
		MirrorID: d.mir.ID.String(),
		Type:     util.MirrorActivityTypeSync,
	})
	if err != nil {
		return nil, err
	}

	flow.pubsub.NotifyMirror(d.ino)

	var resp emptypb.Empty

	return &resp, nil

}

func (flow *flow) HardSyncMirror(ctx context.Context, req *grpc.HardSyncMirrorRequest) (*emptypb.Empty, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	d, err := flow.traverseToMirror(ctx, tx.Namespace, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	err = flow.syncer.NewActivity(tx, &newMirrorActivityArgs{
		MirrorID: d.mir.ID.String(),
		Type:     util.MirrorActivityTypeSync,
	})
	if err != nil {
		return nil, err
	}

	flow.pubsub.NotifyMirror(d.ino)

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

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	d, err := flow.traverseToMirror(ctx, tx.Namespace, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	query := d.mir.QueryActivities()

	results, pi, err := paginate[*ent.MirrorActivityQuery, *ent.MirrorActivity](ctx, req.Pagination, query, mirrorActivitiesOrderings, mirrorActivitiesFilters)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	resp := new(grpc.MirrorInfoResponse)
	resp.Namespace = d.namespace()
	resp.Activities = new(grpc.MirrorActivities)
	resp.Activities.PageInfo = pi

	err = atob(results, &resp.Activities.Results)
	if err != nil {
		return nil, err
	}

	err = atob(d.mir, &resp.Info)
	if err != nil {
		return nil, err
	}

	if d.mir.Passphrase != "" {
		resp.Info.Passphrase = "-"
	}
	if d.mir.PrivateKey != "" {
		resp.Info.PrivateKey = "-"
	}

	return resp, nil

}

func (flow *flow) MirrorInfoStream(req *grpc.MirrorInfoRequest, srv grpc.Flow_MirrorInfoStreamServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	d, err := flow.traverseToMirror(ctx, flow.db.Namespace, req.GetNamespace(), req.GetPath())
	if err != nil {
		return err
	}

	if d.ino.ExtendedType != util.InodeTypeGit {
		return ErrNotMirror
	}

	sub := flow.pubsub.SubscribeMirror(d.ino)
	defer flow.cleanup(sub.Close)

resend:

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx)

	d, err = flow.traverseToMirror(ctx, tx.Namespace, d.namespace(), d.path)
	if err != nil {
		return err
	}

	query := d.mir.QueryActivities()

	results, pi, err := paginate[*ent.MirrorActivityQuery, *ent.MirrorActivity](ctx, req.Pagination, query, mirrorActivitiesOrderings, mirrorActivitiesFilters)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	resp := new(grpc.MirrorInfoResponse)
	resp.Namespace = d.namespace()
	resp.Activities = new(grpc.MirrorActivities)
	resp.Activities.PageInfo = pi

	err = atob(results, &resp.Activities.Results)
	if err != nil {
		return err
	}

	err = atob(d.mir, &resp.Info)
	if err != nil {
		return err
	}

	if d.mir.Passphrase != "" {
		resp.Info.Passphrase = "-"
	}
	if d.mir.PrivateKey != "" {
		resp.Info.PrivateKey = "-"
	}

	nhash = checksum(resp)
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

type mirrorActivityData struct {
	act *ent.MirrorActivity
	// *mirData
}

func (d *mirrorActivityData) namespace() string {
	return d.act.Edges.Namespace.Name
}

func (srv *server) getMirrorActivity(ctx context.Context, nsc *ent.NamespaceClient, namespace, activity string) (*mirrorActivityData, error) {

	id, err := uuid.Parse(activity)
	if err != nil {
		srv.sugar.Debugf("%s failed to parse UUID: %v", parent(), err)
		return nil, err
	}

	ns, err := srv.getNamespace(ctx, nsc, namespace)
	if err != nil {
		srv.sugar.Debugf("%s failed to resolve namespace: %v", parent(), err)
		return nil, err
	}

	query := ns.QueryMirrorActivities().Where(entmiract.IDEQ(id))
	act, err := query.Only(ctx)
	if err != nil {
		srv.sugar.Debugf("%s failed to query instance: %v", parent(), err)
		return nil, err
	}

	act.Edges.Namespace = ns

	d := new(mirrorActivityData)
	d.act = act

	return d, nil

}

func (flow *flow) MirrorActivityLogs(ctx context.Context, req *grpc.MirrorActivityLogsRequest) (*grpc.MirrorActivityLogsResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	d, err := flow.getMirrorActivity(ctx, flow.db.Namespace, req.GetNamespace(), req.GetActivity())
	if err != nil {
		return nil, err
	}

	query := d.act.QueryLogs()

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logsFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.MirrorActivityLogsResponse)
	resp.Namespace = d.namespace()
	resp.Activity = d.act.ID.String()
	resp.PageInfo = pi

	err = atob(results, &resp.Results)
	if err != nil {
		return nil, err
	}

	return resp, nil

}

func (flow *flow) MirrorActivityLogsParcels(req *grpc.MirrorActivityLogsRequest, srv grpc.Flow_MirrorActivityLogsParcelsServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	var tailing bool

	d, err := flow.getMirrorActivity(ctx, flow.db.Namespace, req.GetNamespace(), req.GetActivity())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeMirrorActivityLogs(d.act)
	defer flow.cleanup(sub.Close)

resend:

	query := d.act.QueryLogs()

	results, pi, err := paginate[*ent.LogMsgQuery, *ent.LogMsg](ctx, req.Pagination, query, logsOrderings, logsFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.MirrorActivityLogsResponse)
	resp.Namespace = d.namespace()
	resp.Activity = d.act.ID.String()
	resp.PageInfo = pi

	err = atob(results, &resp.Results)
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

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	flow.syncer.cancelActivity(req.GetActivity(), "cancel.api", "cancelled by api request")

	var resp emptypb.Empty

	return &resp, nil

}
