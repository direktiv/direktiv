package flow

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/ent"
	entact "github.com/direktiv/direktiv/pkg/flow/ent/mirroractivity"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/project"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/gobwas/glob"
	"github.com/google/uuid"
	git "github.com/libgit2/git2go/v33"
	"github.com/mitchellh/hashstructure/v2"
	"gopkg.in/yaml.v3"
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

		if derrors.IsNotFound(err) {
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
		LockCtx:  ctx,
		LockConn: conn,
	})
	if err != nil {
		if err == ErrMirrorLocked {
			return
		}
		syncer.sugar.Error(err)
		return
	}

}

// locks

func (syncer *syncer) lock(key string, timeout time.Duration) (context.Context, *sql.Conn, error) {

	hash, err := hashstructure.Hash(key, hashstructure.FormatV2, nil)
	if err != nil {
		return nil, nil, derrors.NewInternalError(err)
	}

	wait := int(timeout.Seconds())

	conn, err := syncer.locks.lockDB(hash, wait)
	if err != nil {
		return nil, nil, derrors.NewInternalError(err)
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
	LockCtx  context.Context
	LockConn *sql.Conn
}

func (syncer *syncer) beginActivity(tx *ent.Tx, args *newMirrorActivityArgs) (*activityMemory, error) {

	var err error
	var ctx context.Context
	var conn *sql.Conn

	if args.LockConn != nil {
		ctx = args.LockCtx
		conn = args.LockConn
	} else {
		ctx, conn, err = syncer.lock(args.MirrorID, defaultLockWait)
		if err != nil {
			return nil, err
		}
		defer syncer.unlock(args.MirrorID, conn)
	}

	if tx == nil {
		tx, err = syncer.db.Tx(ctx)
		if err != nil {
			return nil, err
		}
		defer rollback(tx)
	}

	d, err := syncer.reverseTraverseToMirror(ctx, tx.Inode, tx.Mirror, args.MirrorID)
	if err != nil {
		return nil, err
	}

	unfinishedActivities, err := d.mir.QueryActivities().Where(entact.StatusIn(util.MirrorActivityStatusPending, util.MirrorActivityStatusExecuting)).Count(ctx)
	if err != nil {
		return nil, err
	}

	if unfinishedActivities > 0 {
		return nil, errors.New("mirror operations are already underway")
	}

	if !d.ino.ReadOnly {
		switch args.Type {
		case util.MirrorActivityTypeLocked:
		default:
			return nil, ErrMirrorLocked
		}
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
		err = syncer.initMirror(ctx, am)
		if err != nil {
			return
		}
	case util.MirrorActivityTypeLocked: // NOTE: intentionally left empty
	case util.MirrorActivityTypeUnlocked: // NOTE: intentionally left empty
	case util.MirrorActivityTypeReconfigure:
		err = syncer.hardSync(ctx, am)
		if err != nil {
			return
		}

		// in case of cron
		syncer.server.syncerCronPollerMirror(am.mir)

	case util.MirrorActivityTypeCronSync:
		err = syncer.hardSync(ctx, am)
		if err != nil {
			return
		}
	case util.MirrorActivityTypeSync:
		err = syncer.hardSync(ctx, am)
		if err != nil {
			return
		}
	default:
		syncer.logToMirrorActivity(ctx, time.Now(), am.act, "Unrecognized syncer activity type.")
	}

	err = syncer.success(am)
	if err != nil {
		return
	}

}

func (syncer *syncer) fail(am *activityMemory, e error) {

	if e == nil {
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

	syncer.logToMirrorActivity(ctx, time.Now(), act, "Mirror activity '%s' failed: %v", act.Type, e)

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

func (syncer *syncer) initMirror(ctx context.Context, am *activityMemory) error {

	return syncer.hardSync(ctx, am)

}

func (syncer *syncer) tarGzDir(path string) ([]byte, error) {

	tf, err := ioutil.TempFile("", "outtar")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tf.Name())

	err = tarGzDir(path, tf)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadFile(tf.Name())

}

func tarGzDir(src string, buf io.Writer) error {

	zr := gzip.NewWriter(buf)
	tw := tar.NewWriter(zr)

	err := filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {

		if !fi.Mode().IsDir() && !fi.Mode().IsRegular() {
			return nil
		}

		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		// use "subpath"
		header.Name = filepath.ToSlash(file[len(src):])

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !fi.IsDir() {
			/* #nosec */
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, data); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	if err := tw.Close(); err != nil {
		return err
	}

	if err := zr.Close(); err != nil {
		return err
	}

	return nil
}

func (syncer *syncer) hardSync(ctx context.Context, am *activityMemory) error {

	lr, err := loadRepository(ctx, &repositorySettings{
		UUID:       am.mir.ID,
		URL:        am.mir.URL,
		Branch:     am.mir.Ref,
		Passphrase: am.mir.Passphrase,
		PrivateKey: am.mir.PrivateKey,
		PublicKey:  am.mir.PublicKey,
	})
	if err != nil {
		return err
	}
	defer lr.Cleanup()

	// lr.lastCommit = opts.LastCommit

	model, err := buildModel(ctx, lr)
	if err != nil {
		return err
	}

	tx, err := syncer.db.Tx(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx)

	md, err := syncer.reverseTraverseToMirror(ctx, tx.Inode, tx.Mirror, am.mir.ID.String())
	if err != nil {
		return err
	}

	cache := make(map[string]*ent.Inode)
	trueroot := filepath.Join(md.path, ".")
	if trueroot == "/" {
		cache[trueroot] = md.ino
	} else {
		cache[trueroot+"/"] = md.ino
	}

	var recurser func(parent *ent.Inode, path string) error
	recurser = func(parent *ent.Inode, path string) error {

		children, err := parent.Children(ctx)
		if err != nil {
			return err
		}

		for _, child := range children {

			cpath := filepath.Join(path, child.Name)
			actualpath := filepath.Join(trueroot, cpath)

			if child.Type == util.InodeTypeDirectory && child.ExtendedType == util.InodeTypeGit {

				_, err = model.lookup(cpath)
				if err == os.ErrNotExist {
					continue
				}
				if err != nil {
					return err
				}

				err = syncer.flow.deleteNode(ctx, &deleteNodeArgs{
					inoc:      tx.Inode,
					ns:        md.ns(),
					pino:      parent,
					ino:       child,
					path:      actualpath,
					super:     true,
					recursive: true,
				})
				if err != nil {
					return err
				}

			} else if child.Type == util.InodeTypeDirectory {

				err = recurser(child, cpath)
				if err != nil {
					return err
				}

				mn, err := model.lookup(cpath)
				if err != nil && err != os.ErrNotExist {
					return err
				}

				if err == os.ErrNotExist || mn.ntype != mntDir {

					err = syncer.flow.deleteNode(ctx, &deleteNodeArgs{
						inoc:      tx.Inode,
						ns:        md.ns(),
						pino:      parent,
						ino:       child,
						path:      actualpath,
						super:     true,
						recursive: true,
					})
					if err != nil {
						return err
					}
				}

			} else if child.Type == util.InodeTypeWorkflow {

				mn, err := model.lookup(cpath)
				if err != nil && err != os.ErrNotExist {
					return err
				}

				if err == os.ErrNotExist || mn.ntype != mntWorkflow {

					err = syncer.flow.deleteNode(ctx, &deleteNodeArgs{
						inoc:      tx.Inode,
						ns:        md.ns(),
						pino:      parent,
						ino:       child,
						path:      actualpath,
						super:     true,
						recursive: true,
					})
					if err != nil {
						return err
					}
				}

			} else {
				return errors.New("how?")
			}
		}

		return nil

	}

	err = recurser(md.ino, ".")
	if err != nil {
		return err
	}

	err = modelWalk(model.root, ".", func(path string, n *mirrorNode, err error) error {

		if path == "." {
			return nil
		}

		truepath := filepath.Join(md.path, path)
		dir, base := filepath.Split(truepath)

		switch n.ntype {
		case mntDir:

			ino, err := syncer.flow.createDirectory(ctx, &createDirectoryArgs{
				inoc:  tx.Inode,
				ns:    md.ns(),
				pino:  cache[dir],
				path:  truepath,
				super: true,
			})
			if ino == nil {
				return err
			}

			cache[truepath+"/"] = ino

		case mntWorkflow:

			data, err := ioutil.ReadFile(filepath.Join(lr.path, path+n.extension))
			if err != nil {
				return err
			}

			wf, err := syncer.flow.createWorkflow(ctx, &createWorkflowArgs{
				inoc: tx.Inode,
				wfc:  tx.Workflow,
				revc: tx.Revision,
				refc: tx.Ref,
				evc:  tx.Events,

				ns:         md.ns(),
				pino:       cache[dir],
				path:       truepath,
				super:      true,
				data:       data,
				noValidate: true,
			})
			if wf == nil {
				return err
			}
			if err == os.ErrExist {
				_, err = syncer.flow.updateWorkflow(ctx, &updateWorkflowArgs{
					revc:       tx.Revision,
					eventc:     tx.Events,
					ns:         md.ns(),
					ino:        wf.Edges.Inode,
					wf:         wf,
					path:       truepath,
					super:      true,
					data:       data,
					noValidate: true,
				})
				if err != nil {
					return err
				}
			}

			cache[truepath+"/"] = wf.Edges.Inode

		case mntNamespaceVar:

			var data []byte
			var err error

			fpath := filepath.Join(lr.path, path)

			if n.isDir {
				data, err = syncer.tarGzDir(fpath)
			} else {
				data, err = ioutil.ReadFile(fpath)
			}
			if err != nil {
				return err
			}

			_, base := filepath.Split(path)
			trimmed := strings.TrimPrefix(base, "var.")

			_, _, err = syncer.flow.SetVariable(ctx, tx.VarRef, tx.VarData, md.ns(), trimmed, data, "", false)
			if err != nil {
				return err
			}

		case mntWorkflowVar:

			var data []byte
			var err error

			fpath := filepath.Join(lr.path, path)

			if n.isDir {
				data, err = syncer.tarGzDir(fpath)
			} else {
				data, err = ioutil.ReadFile(fpath)
			}
			if err != nil {
				return err
			}

			x := strings.SplitN(truepath, ".yaml.", 2)
			if len(x) == 1 {
				x = strings.SplitN(truepath, ".yml.", 2)
				if len(x) == 1 {
					return errors.New("how did this happen?")
				}
			}
			trimmed := x[1]
			_, base = filepath.Split(x[0])

			pino := cache[dir]
			wf, err := syncer.flow.lookupWorkflowFromParent(ctx, &lookupWorkflowFromParentArgs{
				pino: pino,
				name: base,
			})
			if err != nil {
				if derrors.IsNotFound(err) {
					syncer.logToMirrorActivity(ctx, time.Now(), am.act, "Found something that looks like a workflow variable with no matching workflow: "+path)
					return nil
				}
				return err
			}

			_, _, err = syncer.flow.SetVariable(ctx, tx.VarRef, tx.VarData, wf, trimmed, data, "", false)
			if err != nil {
				return err
			}

		default:
			return errors.New("unexpected mnt")
		}

		return nil

	})
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return nil
	}

	return nil

}

type repositorySettings struct {
	UUID   uuid.UUID
	URL    string
	Branch string
	// Username   string
	Passphrase string
	PrivateKey string
	PublicKey  string
}

type localRepository struct {
	path          string
	repo          *repositorySettings
	gitRepository *git.Repository
	lastCommit    string
}

func loadRepository(ctx context.Context, repo *repositorySettings) (*localRepository, error) {

	repository := new(localRepository)
	repository.repo = repo
	repository.path = filepath.Join(os.TempDir(), repo.UUID.String())

	_ = os.RemoveAll(repository.path)

	err := repository.clone(ctx)
	if err != nil {
		return nil, err
	}

	return repository, nil

}

func (repository *localRepository) Cleanup() {
	repository.gitRepository.Free()
	_ = os.RemoveAll(repository.path)
}

func (repository *localRepository) clone(ctx context.Context) error {

	checkoutOpts := git.CheckoutOptions{
		Strategy: git.CheckoutForce,
	}

	fetchOpts := git.FetchOptions{
		RemoteCallbacks: git.RemoteCallbacks{
			CertificateCheckCallback: func(cert *git.Certificate, valid bool, hostname string) error { return nil }, // TODO
			CredentialsCallback: func(url string, username_from_url string, allowed_types git.CredentialType) (*git.Credential, error) {
				cred, err := git.NewCredentialSSHKeyFromMemory(username_from_url, repository.repo.PublicKey, repository.repo.PrivateKey, repository.repo.Passphrase)
				if err != nil {
					fmt.Println(err)
					return nil, err
				}
				return cred, err
			},
		},
	}

	uri := repository.repo.URL
	prefix := "https://"

	if strings.HasPrefix(uri, prefix) && len(repository.repo.Passphrase) > 0 {
		if !strings.Contains(uri, "@") {
			uri = fmt.Sprintf("%s%s@", prefix, url.QueryEscape(repository.repo.Passphrase)) + strings.TrimPrefix(uri, prefix)
		}
	}

	r, err := git.Clone(uri, repository.path, &git.CloneOptions{
		CheckoutOptions: checkoutOpts,
		FetchOptions:    fetchOpts,
		CheckoutBranch:  repository.repo.Branch,
	})

	if err != nil {
		return err
	}

	repository.gitRepository = r

	return nil

}

const (
	mntDir          = "directory"
	mntWorkflow     = "workflow"
	mntWorkflowVar  = "wfvar"
	mntNamespaceVar = "nsvar"
)

var (
	ntypeShort map[string]string
)

func init() {
	ntypeShort = map[string]string{
		mntDir:      "d",
		mntWorkflow: "w",
	}
}

type mirrorNode struct {
	parent    *mirrorNode
	children  []*mirrorNode
	ntype     string
	name      string
	change    string
	extension string
	isDir     bool
}

type mirrorModel struct {
	root *mirrorNode
}

func (model *mirrorModel) lookup(path string) (*mirrorNode, error) {

	path = strings.Trim(path, "/")

	if path == "." || path == "" {
		return model.root, nil
	}

	dir, base := filepath.Split(path)

	if base == "" {
		base = dir
		dir = "."
	}

	pnode, err := model.lookup(dir)
	if err != nil {
		return nil, err
	}

	for _, c := range pnode.children {
		if c.name == base {
			return c, nil
		}
	}

	return nil, os.ErrNotExist

}

func (model *mirrorModel) addDirectoryNode(path string) error {

	dir, base := filepath.Split(path)

	node, err := model.lookup(dir)
	if err != nil {
		return err
	}

	if node.ntype != mntDir {
		return errors.New("parent not a directory")
	}

	node.children = append(node.children, &mirrorNode{
		parent:   node,
		children: make([]*mirrorNode, 0),
		ntype:    mntDir,
		name:     base,
	})

	return nil

}

func (model *mirrorModel) addWorkflowNode(path string) error {

	dir, base := filepath.Split(path)

	node, err := model.lookup(dir)
	if err != nil {
		return err
	}

	if node.ntype != mntDir {
		return errors.New("parent not a directory")
	}

	var name, extension string

	if strings.HasSuffix(base, ".yml") {
		name = base[:len(base)-4]
		extension = ".yml"
	} else if strings.HasSuffix(base, ".yaml") {
		name = base[:len(base)-5]
		extension = ".yaml"
	}

	if !util.NameRegex.MatchString(name) {
		// TODO: log something
		return nil
	}

	node.children = append(node.children, &mirrorNode{
		parent:    node,
		children:  make([]*mirrorNode, 0),
		ntype:     mntWorkflow,
		name:      name,
		extension: extension,
	})

	return nil

}

func (model *mirrorModel) addWorkflowVariableNode(path string, isDir bool) error {

	dir, base := filepath.Split(path)

	node, err := model.lookup(dir)
	if err != nil {
		return err
	}

	if node.ntype != mntDir {
		return errors.New("parent not a directory")
	}

	node.children = append(node.children, &mirrorNode{
		parent:   node,
		children: make([]*mirrorNode, 0),
		ntype:    mntWorkflowVar,
		name:     base,
		isDir:    isDir,
	})

	return nil

}

func (model *mirrorModel) addNamespaceVariableNode(path string, isDir bool) error {

	dir, base := filepath.Split(path)

	node, err := model.lookup(dir)
	if err != nil {
		return err
	}

	if node.ntype != mntDir {
		return errors.New("parent not a directory")
	}

	node.children = append(node.children, &mirrorNode{
		parent:   node,
		children: make([]*mirrorNode, 0),
		ntype:    mntNamespaceVar,
		name:     base,
		isDir:    isDir,
	})

	return nil

}

func (model *mirrorModel) finalize() error {

	var recurse func(node *mirrorNode) bool

	recurse = func(node *mirrorNode) bool {

		if node.ntype != mntDir {
			return true
		}

		children := make([]*mirrorNode, 0)

		for _, c := range node.children {
			hasData := recurse(c)
			if hasData {
				children = append(children, c)
			}
		}

		node.children = children
		return len(node.children) > 0

	}

	recurse(model.root)

	return nil

}

func (model *mirrorModel) dump() {

	err := modelWalk(model.root, ".", func(path string, n *mirrorNode, err error) error {
		fmt.Println(ntypeShort[n.ntype], path, n.change)
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}

}

func modelWalk(node *mirrorNode, path string, fn func(path string, n *mirrorNode, err error) error) error {

	err := fn(path, node, nil)
	if err == filepath.SkipDir {
		return nil
	}
	if err != nil {
		return err
	}

	for _, c := range node.children {
		cpath := filepath.Join(path, c.name)
		err := modelWalk(c, cpath, fn)
		if err != nil {
			return err
		}
	}

	return nil

}

func (model *mirrorModel) diff(repo *localRepository) error {

	var oldid *git.Oid

	oldref, err := repo.gitRepository.References.Lookup(repo.lastCommit)
	if err != nil {

		oldid, err = git.NewOid(repo.lastCommit)
		if err != nil {
			return err
		}

	} else {

		defer oldref.Free()
		oldid = oldref.Target()

		if oldid == nil {
			href, err := repo.gitRepository.Head()
			if err != nil {
				return err
			}
			defer href.Free()
			oldid = href.Target()
		}

	}

	oldcommit, err := repo.gitRepository.LookupCommit(oldid)
	if err != nil {
		return err
	}
	defer oldcommit.Free()

	oldtree, err := oldcommit.Tree()
	if err != nil {
		return err
	}
	defer oldtree.Free()

	href, err := repo.gitRepository.Head()
	if err != nil {
		return err
	}
	defer href.Free()
	newid := href.Target()

	newcommit, err := repo.gitRepository.LookupCommit(newid)
	if err != nil {
		return err
	}
	defer newcommit.Free()

	newtree, err := newcommit.Tree()
	if err != nil {
		return err
	}
	defer newtree.Free()

	opts, err := git.DefaultDiffOptions()
	if err != nil {
		return err
	}

	diff, err := repo.gitRepository.DiffTreeToTree(oldtree, newtree, &opts)
	if err != nil {
		return err
	}
	defer diff.Free()

	dopts, err := git.DefaultDiffFindOptions()
	if err != nil {
		return err
	}

	err = diff.FindSimilar(&dopts)
	if err != nil {
		return err
	}

	err = diff.ForEach(func(delta git.DiffDelta, progress float64) (git.DiffForEachHunkCallback, error) {

		node, err := model.lookup(delta.NewFile.Path)
		if err == os.ErrNotExist {
			return func(x git.DiffHunk) (git.DiffForEachLineCallback, error) {
				return func(x git.DiffLine) error {
					return nil
				}, nil
			}, nil
		}
		if err != nil {
			return nil, err
		}

		switch delta.Status {
		case git.DeltaAdded:
			node.change = "added"
		case git.DeltaModified:
			node.change = "changed"
		case git.DeltaRenamed:
			node.change = fmt.Sprintf("renamed from '%s'", delta.OldFile.Path)
		default:
			node.change = delta.Status.String()
		}

		return func(x git.DiffHunk) (git.DiffForEachLineCallback, error) {
			return func(x git.DiffLine) error {
				return nil
			}, nil
		}, nil
	}, git.DiffDetailFiles)
	if err != nil {
		return err
	}

	return nil

}

func buildModel(ctx context.Context, repo *localRepository) (*mirrorModel, error) {

	model := new(mirrorModel)
	model.root = &mirrorNode{
		parent:   model.root,
		children: make([]*mirrorNode, 0),
		ntype:    mntDir,
		name:     ".", // TODO
	}

	cfg := new(project.Config)

	scfpath := filepath.Join(repo.path, project.ConfigFile)
	scfgbytes, err := ioutil.ReadFile(scfpath)
	if os.IsNotExist(err) {
		cfg.Ignore = make([]string, 0)
	} else if err != nil {
		return nil, fmt.Errorf("failed to read direktiv config file: %w", err)
	} else {
		err := yaml.Unmarshal(scfgbytes, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal direktiv config file: %w", err)
		}
	}

	globbers := make([]glob.Glob, 0)
	for idx, pattern := range cfg.Ignore {
		g, err := glob.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %dth ignore pattern: %w", idx, err)
		}
		globbers = append(globbers, g)
	}

	err = filepath.WalkDir(repo.path, func(path string, d fs.DirEntry, err error) error {

		rel, err := filepath.Rel(repo.path, path)
		if err != nil {
			return err
		}

		if rel == "." {
			return nil
		}

		if rel == ".git" {
			return filepath.SkipDir
		}

		if rel == ".direktiv.yaml" {
			return nil
		}

		for _, g := range globbers {
			if g.Match(rel) {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		_, base := filepath.Split(path)

		if strings.Contains(base, ".yaml.") || strings.Contains(base, ".yml.") {
			err = model.addWorkflowVariableNode(rel, d.IsDir())
			if err != nil {
				return err
			}
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.HasPrefix(base, "var.") {
			err = model.addNamespaceVariableNode(rel, d.IsDir())
			if err != nil {
				return err
			}
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !d.IsDir() && (strings.HasSuffix(base, ".yaml") || strings.HasSuffix(base, ".yml")) {
			err = model.addWorkflowNode(rel)
			if err != nil {
				return err
			}
			return nil
		}

		if d.IsDir() {
			if !util.NameRegex.MatchString(base) {
				return filepath.SkipDir
			}

			err = model.addDirectoryNode(rel)
			if err != nil {
				return err
			}

			return nil
		}

		return nil

	})
	if err != nil {
		return nil, err
	}

	err = model.finalize()
	if err != nil {
		return nil, err
	}

	if repo.lastCommit != "" {
		err = model.diff(repo)
		if err != nil {
			return nil, err
		}
	}

	// model.dump()

	return model, nil

}
