package flow

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/ent"
	entact "github.com/direktiv/direktiv/pkg/flow/ent/mirroractivity"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
	"github.com/mitchellh/hashstructure/v2"
)

type syncer struct {
	*server
	cancellers     map[string]func()
	cancellersLock sync.Mutex
}

func initSyncer(srv *server) (*syncer, error) {

	syncer := new(syncer)

	syncer.server = srv

	syncer.cancellers = make(map[string]func())

	return syncer, nil

}

func (syncer *syncer) Close() error {

	return nil

}

func (srv *server) reverseTraverseToMirror(ctx context.Context, inoc *ent.InodeClient, mirc *ent.MirrorClient, id string) (*mirData, error) {

	uid, err := uuid.Parse(id)
	if err != nil {
		srv.sugar.Debugf("%s failed to parse mirror UUID: %v", parent(), err)
		return nil, err
	}

	mir, err := mirc.Get(ctx, uid)
	if err != nil {
		srv.sugar.Debugf("%s failed to query mirror: %v", parent(), err)
		return nil, err
	}

	ino, err := mir.Inode(ctx)
	if err != nil {
		srv.sugar.Debugf("%s failed to query mirror's inode: %v", parent(), err)
		return nil, err
	}

	nd, err := srv.reverseTraverseToInode(ctx, inoc, ino.ID.String())
	if err != nil {
		srv.sugar.Debugf("%s failed to resolve inode's parent(s): %v", parent(), err)
		return nil, err
	}

	mir.Edges.Inode = nd.ino
	mir.Edges.Namespace = nd.ino.Edges.Namespace

	d := new(mirData)
	d.mir = mir
	d.nodeData = nd

	return d, nil

}

// Timeouts

func (syncer *syncer) scheduleTimeout(activityId string, oldController string, t time.Time) {

	var err error
	deadline := t

	id := fmt.Sprintf("syncertimeout:%s", activityId)

	// cancel existing timeouts

	syncer.timers.deleteTimerByName(oldController, syncer.pubsub.hostname, id)

	// schedule timeout

	args := &syncerTimeoutArgs{
		ActivityId: activityId,
	}

	data, err := json.Marshal(args)
	if err != nil {
		panic(err)
	}

	err = syncer.timers.addOneShot(id, syncerTimeoutFunction, deadline, data)
	if err != nil {
		syncer.sugar.Error(err)
		// TODO: abort?
	}

}

func (syncer *syncer) ScheduleTimeout(activityId, oldController string, t time.Time) {
	syncer.scheduleTimeout(activityId, oldController, t)
}

type syncerTimeoutArgs struct {
	ActivityId string
}

const syncerTimeoutFunction = "syncerTimeoutFunction"

func (syncer *syncer) cancelActivity(activityId, code, message string) {

	am, err := syncer.loadActivityMemory(activityId)
	if err != nil {
		syncer.sugar.Error(err)
		return
	}

	syncer.fail(am, errors.New(code))

}

func (syncer *syncer) timeoutHandler(input []byte) {

	args := new(syncerTimeoutArgs)
	err := json.Unmarshal(input, args)
	if err != nil {
		syncer.sugar.Error(err)
		return
	}

	syncer.cancelActivity(args.ActivityId, ErrCodeSoftTimeout, "syncer activity timed out")

}

// Pollers

func (srv *server) syncerCronPoller() {

	for {
		srv.syncerCronPoll()
		time.Sleep(time.Minute * 15)
	}

}

func (srv *server) syncerCronPoll() {

	ctx := context.Background()

	mirs, err := srv.db.Mirror.Query().All(ctx)
	if err != nil {
		srv.sugar.Error(err)
		return
	}

	for _, mir := range mirs {
		srv.syncerCronPollerMirror(mir)
	}

}

func (srv *server) syncerCronPollerMirror(mir *ent.Mirror) {

	if mir.Cron != "" {
		srv.timers.deleteCronForSyncer(mir.ID.String())

		err := srv.timers.addCron(mir.ID.String(), syncerCron, mir.Cron, []byte(mir.ID.String()))
		if err != nil {
			srv.sugar.Error(err)
			return
		}

		srv.sugar.Debugf("Loaded syncer cron: %s", mir.ID.String())

	}

}

func (syncer *syncer) cronHandler(data []byte) {

	id := string(data)

	ctx, conn, err := syncer.lock(id, defaultLockWait)
	if err != nil {
		syncer.sugar.Error(err)
		return
	}
	defer syncer.unlock(id, conn)

	d, err := syncer.reverseTraverseToMirror(ctx, syncer.db.Inode, syncer.db.Mirror, id)
	if err != nil {

		if IsNotFound(err) {
			syncer.sugar.Infof("Cron failed to find mirror. Deleting cron.")
			syncer.timers.deleteCronForSyncer(id)
			return
		}

		syncer.sugar.Error(err)
		return

	}

	k, err := d.mir.QueryActivities().Where(entact.CreatedAtGT(time.Now().Add(-time.Second*30)), entact.TypeEQ(util.MirrorActivityTypeCronSync)).Count(ctx)
	if err != nil {
		syncer.sugar.Error(err)
		return
	}

	if k != 0 {
		// already triggered
		return
	}

	args := new(newInstanceArgs)
	args.Namespace = d.namespace()
	args.Path = d.path
	args.Ref = ""
	args.Input = nil
	args.Caller = "cron"
	args.CallerData = "cron"

	err = syncer.NewActivity(nil, &newMirrorActivityArgs{
		MirrorID: d.mir.ID.String(),
		Type:     util.MirrorActivityTypeCronSync,
	})
	if err != nil {
		syncer.sugar.Error(err)
		return
	}

}

// locks

func (syncer *syncer) lock(key string, timeout time.Duration) (context.Context, *sql.Conn, error) {

	hash, err := hashstructure.Hash(key, hashstructure.FormatV2, nil)
	if err != nil {
		return nil, nil, NewInternalError(err)
	}

	wait := int(timeout.Seconds())

	conn, err := syncer.locks.lockDB(hash, wait)
	if err != nil {
		return nil, nil, NewInternalError(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	syncer.cancellersLock.Lock()
	syncer.cancellers[key] = cancel
	syncer.cancellersLock.Unlock()

	return ctx, conn, nil

}

func (syncer *syncer) unlock(key string, conn *sql.Conn) {

	hash, err := hashstructure.Hash(key, hashstructure.FormatV2, nil)
	if err != nil {
		panic(err)
	}

	syncer.cancellersLock.Lock()
	defer syncer.cancellersLock.Unlock()

	cancel := syncer.cancellers[key]
	delete(syncer.cancellers, key)
	cancel()

	err = syncer.locks.unlockDB(hash, conn)
	if err != nil {
		syncer.sugar.Error(err)
		return
	}

}

func (syncer *syncer) kickExpiredActivities() {

	ctx := context.Background()

	t := time.Now().Add(-1 * time.Minute)

	list, err := syncer.db.MirrorActivity.Query().
		Where(entact.DeadlineLT(t), entact.StatusIn(util.MirrorActivityStatusExecuting, util.MirrorActivityStatusPending)).
		All(ctx)
	if err != nil {
		syncer.sugar.Error(err)
		return
	}

	for _, act := range list {

		syncer.cancelActivity(act.ID.String(), "timeouts.deadline.exceeded", "Activity failed to terminate before deadline.")

	}

}

// activity memory

type activityMemory struct {
	act *ent.MirrorActivity
	mir *ent.Mirror
	ino *ent.Inode
	ns  *ent.Namespace
}

func (syncer *syncer) loadActivityMemory(id string) (*activityMemory, error) {

	ctx := context.Background()

	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	act, err := syncer.db.MirrorActivity.Get(ctx, uid)
	if err != nil {
		return nil, err
	}

	mir, err := act.QueryMirror().WithNamespace().WithInode().Only(ctx)
	if err != nil {
		return nil, err
	}

	ino, err := mir.QueryInode().Only(ctx)
	if err != nil {
		return nil, err
	}

	am := new(activityMemory)
	am.act = act
	am.mir = mir
	am.ino = ino
	ino.Edges.Namespace = mir.Edges.Namespace
	act.Edges.Namespace = mir.Edges.Namespace
	act.Edges.Mirror = am.mir
	am.ns = act.Edges.Namespace

	return am, nil

}

func (am *activityMemory) ID() uuid.UUID {

	return am.act.ID

}

// activity

type newMirrorActivityArgs struct {
	MirrorID string
	Type     string
}

func (syncer *syncer) beginActivity(tx *ent.Tx, args *newMirrorActivityArgs) (*activityMemory, error) {

	ctx, conn, err := syncer.lock(args.MirrorID, defaultLockWait)
	if err != nil {
		return nil, err
	}
	defer syncer.unlock(args.MirrorID, conn)

	if tx == nil {
		tx, err := syncer.db.Tx(ctx)
		if err != nil {
			return nil, err
		}
		defer rollback(tx)
	}

	d, err := syncer.reverseTraverseToMirror(ctx, tx.Inode, tx.Mirror, args.MirrorID)
	if err != nil {
		return nil, err
	}

	unfinishedActivities, err := d.mir.QueryActivities().Where().Count(ctx)
	if err != nil {
		return nil, err
	}

	if unfinishedActivities > 0 {
		return nil, errors.New("mirror operations are already underway")
	}

	deadline := time.Now().Add(time.Minute * 20)

	act, err := tx.MirrorActivity.Create().
		SetType(args.Type).
		SetStatus(util.MirrorActivityStatusPending).
		SetEndAt(time.Now()).
		SetMirror(d.mir).
		SetNamespace(d.ns()).
		SetController(syncer.pubsub.hostname).
		SetDeadline(deadline).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	act.Edges.Mirror = d.mir
	act.Edges.Namespace = d.ns()

	am := new(activityMemory)
	am.act = act
	am.mir = d.mir
	am.ino = d.ino
	act.Edges.Mirror = am.mir
	act.Edges.Namespace = d.ns()
	am.ns = d.ns()

	syncer.logToNamespace(ctx, time.Now(), d.ns(), "Commenced new mirror activity '%s' on mirror: %s", args.Type, d.path)

	syncer.pubsub.NotifyMirror(d.ino)

	syncer.logToMirrorActivity(ctx, time.Now(), act, "Commenced new mirror activity '%s' on mirror: %s", args.Type, d.path)

	// schedule timeouts
	syncer.scheduleTimeout(am.act.ID.String(), am.act.Controller, deadline)

	return am, nil

}

func (syncer *syncer) NewActivity(tx *ent.Tx, args *newMirrorActivityArgs) error {

	syncer.sugar.Debugf("Handling mirror activity: %s", this())

	am, err := syncer.beginActivity(tx, args)
	if err != nil {
		return err
	}

	go syncer.execute(am)

	return nil

}

func (syncer *syncer) execute(am *activityMemory) {

	var err error

	defer func() {
		syncer.fail(am, err)
	}()

	ctx := context.Background()

	switch am.act.Type {
	case util.MirrorActivityTypeInit:
	case util.MirrorActivityTypeLocked: // NOTE: intentionally left empty
	case util.MirrorActivityTypeUnlocked: // NOTE: intentionally left empty
	case util.MirrorActivityTypeReconfigure:
	case util.MirrorActivityTypeCronSync:
	case util.MirrorActivityTypeSync:
	default:
		syncer.logToMirrorActivity(ctx, time.Now(), am.act, "Unrecognized syncer activity type.")
	}

	err = syncer.success(am)
	if err != nil {
		return
	}

}

func (syncer *syncer) fail(am *activityMemory, err error) {

	if err == nil {
		return
	}

	ctx, conn, err := syncer.lock(am.act.Edges.Mirror.ID.String(), defaultLockWait)
	if err != nil {
		syncer.sugar.Error(err)
		return
	}
	defer syncer.unlock(am.act.Edges.Mirror.ID.String(), conn)

	tx, err := syncer.db.Tx(ctx)
	if err != nil {
		syncer.sugar.Error(err)
		return
	}
	defer rollback(tx)

	edges := am.act.Edges

	act, err := tx.MirrorActivity.Get(ctx, am.act.ID)
	if err != nil {
		syncer.sugar.Error(err)
		return
	}

	if act.Status != util.MirrorActivityStatusExecuting && act.Status != util.MirrorActivityStatusPending {
		err = errors.New("activity somehow already done")
		syncer.sugar.Error(err)
		return
	}

	act, err = act.Update().SetController(syncer.pubsub.hostname).SetEndAt(time.Now()).SetStatus(util.MirrorActivityStatusFailed).Save(ctx)
	if err != nil {
		syncer.sugar.Error(err)
		return
	}

	act.Edges = edges

	err = tx.Commit()
	if err != nil {
		syncer.sugar.Error(err)
		return
	}

	syncer.pubsub.NotifyMirror(am.ino)

	syncer.logToMirrorActivity(ctx, time.Now(), act, "Mirror activity '%s' failed.", act.Type)

	syncer.timers.deleteTimersForActivity(am.ID().String())

}

func (syncer *syncer) success(am *activityMemory) error {

	ctx, conn, err := syncer.lock(am.act.Edges.Mirror.ID.String(), defaultLockWait)
	if err != nil {
		return err
	}
	defer syncer.unlock(am.act.Edges.Mirror.ID.String(), conn)

	tx, err := syncer.db.Tx(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx)

	edges := am.act.Edges

	act, err := tx.MirrorActivity.Get(ctx, am.act.ID)
	if err != nil {
		return err
	}

	if act.Status != util.MirrorActivityStatusExecuting && act.Status != util.MirrorActivityStatusPending {
		return errors.New("activity somehow already done")
	}

	act, err = act.Update().SetController(syncer.pubsub.hostname).SetEndAt(time.Now()).SetStatus(util.MirrorActivityStatusComplete).Save(ctx)
	if err != nil {
		return err
	}

	act.Edges = edges

	err = tx.Commit()
	if err != nil {
		return err
	}

	syncer.pubsub.NotifyMirror(am.ino)

	syncer.logToMirrorActivity(ctx, time.Now(), act, "Completed mirror activity '%s'.", act.Type)

	syncer.timers.deleteTimersForActivity(am.ID().String())

	return nil

}
