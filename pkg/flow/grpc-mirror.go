package flow

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/ent"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	// entmir "github.com/direktiv/direktiv/pkg/flow/ent/mirror"
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
	actc := tx.MirrorActivity
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
		if !IsNotFound(err) {
			return nil, err
		}
	}

	ns, err = nsc.Create().SetName(req.GetName()).Save(ctx)
	if err != nil {
		return nil, err
	}

	ino, err = inoc.Create().SetNillableName(nil).SetType("directory").SetExtendedType(util.InodeTypeGit).SetNamespace(ns).Save(ctx)
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
		SetLocked(false).
		SetNillableLastSync(nil).
		SetInode(ino).
		SetNamespace(ns).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	_, err = actc.Create().
		SetType(util.MirrorActivityTypeInit).
		SetStatus(util.MirrorActivityStatusComplete).
		SetEndAt(time.Now()).
		SetMirror(mir).
		SetNamespace(ns).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
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

	if pino.ino.Type != "directory" {
		return nil, status.Error(codes.AlreadyExists, "parent node is not a directory")
	}

	if pino.ro {
		return nil, errors.New("cannot write into read-only directory")
	}

	settings := req.GetSettings()
	mirc := tx.Mirror
	actc := tx.MirrorActivity
	var mir *ent.Mirror

	ino, err := inoc.Create().SetName(base).SetNamespace(ns).SetParent(pino.ino).SetType("directory").SetExtendedType(util.InodeTypeGit).Save(ctx)
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
		SetLocked(false).
		SetNillableLastSync(nil).
		SetInode(ino).
		SetNamespace(ns).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	_, err = actc.Create().
		SetType(util.MirrorActivityTypeInit).
		SetStatus(util.MirrorActivityStatusComplete).
		SetEndAt(time.Now()).
		SetMirror(mir).
		SetNamespace(ns).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
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

	actc := tx.MirrorActivity

	if !d.mir.Locked {
		_, err = actc.Create().
			SetType(util.MirrorActivityTypeReconfigure).
			SetStatus(util.MirrorActivityStatusComplete).
			SetEndAt(time.Now()).
			SetMirror(d.mir).
			SetNamespace(d.ns()).
			Save(ctx)
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit()
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

	if d.mir.Locked {
		return nil, ErrMirrorLocked
	}

	edges := d.mir.Edges

	d.mir, err = d.mir.Update().SetLocked(true).Save(ctx)
	if err != nil {
		return nil, err
	}

	d.mir.Edges = edges

	actc := tx.MirrorActivity
	_, err = actc.Create().
		SetType(util.MirrorActivityTypeLocked).
		SetStatus(util.MirrorActivityStatusComplete).
		SetEndAt(time.Now()).
		SetMirror(d.mir).
		SetNamespace(d.ns()).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	// flow.logToNamespace(ctx, time.Now(), ns, "Created directory as git mirror '%s'.", path)
	flow.pubsub.NotifyMirror(d.ino)

	// respond:

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

	if !d.mir.Locked {
		return nil, ErrMirrorUnlocked
	}

	edges := d.mir.Edges

	d.mir, err = d.mir.Update().SetLocked(false).Save(ctx)
	if err != nil {
		return nil, err
	}

	d.mir.Edges = edges

	actc := tx.MirrorActivity
	_, err = actc.Create().
		SetType(util.MirrorActivityTypeUnlocked).
		SetStatus(util.MirrorActivityStatusComplete).
		SetEndAt(time.Now()).
		SetMirror(d.mir).
		SetNamespace(d.ns()).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	// flow.logToNamespace(ctx, time.Now(), ns, "Created directory as git mirror '%s'.", path)
	flow.pubsub.NotifyMirror(d.ino)

	// respond:

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

	_, err = flow.traverseToMirror(ctx, tx.Namespace, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	// flow.logToNamespace(ctx, time.Now(), ns, "Created directory as git mirror '%s'.", path)
	// flow.pubsub.NotifyInode(pino.ino)

	// respond:

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

	_, err = flow.traverseToMirror(ctx, tx.Namespace, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	// flow.logToNamespace(ctx, time.Now(), ns, "Created directory as git mirror '%s'.", path)
	// flow.pubsub.NotifyInode(pino.ino)

	// respond:

	var resp emptypb.Empty

	return &resp, nil

}

func mirrorActivityOrder(p *pagination) []ent.MirrorActivityPaginateOption {

	var opts []ent.MirrorActivityPaginateOption

	for _, o := range p.order {

		if o == nil {
			continue
		}

		order := ent.MirrorActivityOrder{
			Direction: ent.OrderDirectionAsc,
			Field:     ent.MirrorActivityOrderFieldCreatedAt,
		}

		switch o.GetField() {
		case "CREATED":
			order.Field = ent.MirrorActivityOrderFieldCreatedAt
		default:
			break
		}

		switch o.GetDirection() {
		case "DESC":
			order.Direction = ent.OrderDirectionDesc
		case "ASC":
			order.Direction = ent.OrderDirectionAsc
		default:
			break
		}

		opts = append(opts, ent.WithMirrorActivityOrder(&order))

	}

	if len(opts) == 0 {
		opts = append(opts, ent.WithMirrorActivityOrder(&ent.MirrorActivityOrder{
			Direction: ent.OrderDirectionAsc,
			Field:     ent.MirrorActivityOrderFieldCreatedAt,
		}))
	}

	return opts

}

func mirrorActivityFilter(p *pagination) []ent.MirrorActivityPaginateOption {

	var filters []func(query *ent.MirrorActivityQuery) (*ent.MirrorActivityQuery, error)
	var opts []ent.MirrorActivityPaginateOption

	if p.filter == nil {
		return nil
	}

	for i := range p.filter {

		f := p.filter[i]

		if f == nil {
			continue
		}

		// filter := f.Val

		filters = append(filters, func(query *ent.MirrorActivityQuery) (*ent.MirrorActivityQuery, error) {

			return query, nil

		})

	}

	if len(filters) > 0 {
		opts = append(opts, ent.WithMirrorActivityFilter(func(query *ent.MirrorActivityQuery) (*ent.MirrorActivityQuery, error) {
			var err error
			for _, filter := range filters {
				query, err = filter(query)
				if err != nil {
					return nil, err
				}
			}
			return query, nil
		}))
	}

	return opts

}

func (flow *flow) MirrorInfo(ctx context.Context, req *grpc.MirrorInfoRequest) (*grpc.MirrorInfoResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	p, err := getPagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	opts := []ent.MirrorActivityPaginateOption{}
	opts = append(opts, mirrorActivityOrder(p)...)
	opts = append(opts, mirrorActivityFilter(p)...)

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
	cx, err := query.Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	// flow.logToNamespace(ctx, time.Now(), ns, "Created directory as git mirror '%s'.", path)
	// flow.pubsub.NotifyInode(pino.ino)

	// respond:

	var resp grpc.MirrorInfoResponse

	err = atob(d.mir, &resp.Info)
	if err != nil {
		return nil, err
	}

	resp.Namespace = d.ns().Name

	err = atob(cx, &resp.Activities)
	if err != nil {
		return nil, err
	}

	return &resp, nil

}

func (flow *flow) MirrorInfoStream(req *grpc.MirrorInfoRequest, srv grpc.Flow_MirrorInfoStreamServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	p, err := getPagination(req.Pagination)
	if err != nil {
		return err
	}

	opts := []ent.MirrorActivityPaginateOption{}
	opts = append(opts, mirrorActivityOrder(p)...)
	opts = append(opts, mirrorActivityFilter(p)...)

	nsc := flow.db.Namespace
	d, err := flow.traverseToMirror(ctx, nsc, req.GetNamespace(), req.GetPath())
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

	d, err = flow.traverseToMirror(ctx, tx.Namespace, req.GetNamespace(), req.GetPath())
	if err != nil {
		return err
	}

	query := d.mir.QueryActivities()
	cx, err := query.Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	// flow.logToNamespace(ctx, time.Now(), ns, "Created directory as git mirror '%s'.", path)
	// flow.pubsub.NotifyInode(pino.ino)

	// respond:

	resp := new(grpc.MirrorInfoResponse)

	err = atob(d.mir, &resp.Info)
	if err != nil {
		return err
	}

	resp.Namespace = d.ns().Name

	err = atob(cx, &resp.Activities)
	if err != nil {
		return err
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
	*mirData
}

func (d *mirrorActivityData) ns() *ent.Namespace {
	return d.act.Edges.Namespace
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

	p, err := getPagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	opts := []ent.LogMsgPaginateOption{}
	opts = append(opts, logsOrder(p)...)
	opts = append(opts, logsFilter(p)...)

	nsc := flow.db.Namespace
	d, err := flow.getMirrorActivity(ctx, nsc, req.GetNamespace(), req.GetActivity())
	if err != nil {
		return nil, err
	}

	cx, err := d.act.QueryLogs().Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return nil, err
	}

	var resp grpc.MirrorActivityLogsResponse

	err = atob(cx, &resp)
	if err != nil {
		return nil, err
	}

	resp.Namespace = d.namespace()
	resp.Activity = d.act.ID.String()

	return &resp, nil

}

func (flow *flow) MirrorActivityLogsParcels(req *grpc.MirrorActivityLogsRequest, srv grpc.Flow_MirrorActivityLogsParcelsServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	var tailing bool

	p, err := getPagination(req.Pagination)
	if err != nil {
		return err
	}

	porder := p.order
	pfilter := p.filter

	opts := []ent.LogMsgPaginateOption{}
	opts = append(opts, logsOrder(p)...)
	opts = append(opts, logsFilter(p)...)

	nsc := flow.db.Namespace
	d, err := flow.getMirrorActivity(ctx, nsc, req.GetNamespace(), req.GetActivity())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeMirrorActivityLogs(d.act)
	defer flow.cleanup(sub.Close)

resend:

	cx, err := d.act.QueryLogs().Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return err
	}

	var resp = new(grpc.MirrorActivityLogsResponse)

	err = atob(cx, resp)
	if err != nil {
		return err
	}

	resp.Namespace = d.namespace()
	resp.Activity = d.act.ID.String()

	if len(resp.Edges) != 0 || !tailing {

		tailing = true

		err = srv.Send(resp)
		if err != nil {
			return err
		}

		p = new(pagination)
		p.after = resp.PageInfo.EndCursor
		p.order = porder
		p.filter = pfilter

	}

	more := sub.Wait(ctx)
	if !more {
		return nil
	}

	goto resend

}
