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
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	gitSSH "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"golang.org/x/crypto/ssh"

	"github.com/direktiv/direktiv/pkg/flow/database"
	entmir "github.com/direktiv/direktiv/pkg/flow/ent/mirror"
	entact "github.com/direktiv/direktiv/pkg/flow/ent/mirroractivity"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/project"
	"github.com/direktiv/direktiv/pkg/util"
	git "github.com/go-git/go-git/v5"
	"github.com/gobwas/glob"
	"github.com/google/uuid"
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

func (srv *server) reverseTraverseToMirror(ctx context.Context, id string) (*database.CacheData, *database.Mirror, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		srv.sugar.Debugf("%s failed to parse mirror UUID: %v", parent(), err)
		return nil, nil, err
	}

	mirror, err := srv.database.Mirror(ctx, uid)
	if err != nil {
		srv.sugar.Debugf("%s failed to query mirror: %v", parent(), err)
		return nil, nil, err
	}

	cached := new(database.CacheData)

	err = srv.database.Inode(ctx, cached, mirror.Inode)
	if err != nil {
		srv.sugar.Debugf("%s failed to resolve inode's parent(s): %v", parent(), err)
		return nil, nil, err
	}

	return cached, mirror, nil
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
	cached, mirror, activity, err := syncer.loadActivityMemory(activityId)
	if err != nil {
		syncer.sugar.Error(err)
		return
	}

	syncer.fail(cached, mirror, activity, errors.New(code))
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

	ids, err := srv.database.Mirrors(ctx)
	if err != nil {
		srv.sugar.Error(err)
		return
	}

	for _, id := range ids {

		mirror, err := srv.database.Mirror(ctx, id)
		if err != nil {
			srv.sugar.Error(err)
			return
		}

		srv.syncerCronPollerMirror(mirror)

	}
}

func (srv *server) syncerCronPollerMirror(mir *database.Mirror) {
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

	cached, mirror, err := syncer.reverseTraverseToMirror(ctx, id)
	if err != nil {

		if derrors.IsNotFound(err) {
			syncer.sugar.Infof("Cron failed to find mirror. Deleting cron.")
			syncer.timers.deleteCronForSyncer(id)
			return
		}

		syncer.sugar.Error(err)
		return

	}

	clients := syncer.edb.Clients(ctx)

	k, err := clients.MirrorActivity.Query().Where(entact.HasMirrorWith(entmir.ID(mirror.ID))).Where(entact.CreatedAtGT(time.Now().Add(-time.Second*30)), entact.TypeEQ(util.MirrorActivityTypeCronSync)).Count(ctx)
	if err != nil {
		syncer.sugar.Error(err)
		return
	}

	if k != 0 {
		// already triggered
		return
	}

	args := new(newInstanceArgs)
	args.Namespace = cached.Namespace.Name
	args.Path = cached.Path()
	args.Ref = ""
	args.Input = nil
	args.Caller = util.CallerCron
	args.CallerData = util.CallerCron

	err = syncer.NewActivity(nil, &newMirrorActivityArgs{
		MirrorID: mirror.ID.String(),
		Type:     util.MirrorActivityTypeCronSync,
		LockCtx:  ctx,
		LockConn: conn,
	})
	if err != nil {
		if errors.Is(err, ErrMirrorLocked) {
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

	clients := syncer.edb.Clients(ctx)

	list, err := clients.MirrorActivity.Query().
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

func (syncer *syncer) loadActivityMemory(id string) (*database.CacheData, *database.Mirror, *database.MirrorActivity, error) {
	ctx := context.Background()

	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, nil, nil, err
	}

	act, err := syncer.database.MirrorActivity(ctx, uid)
	if err != nil {
		return nil, nil, nil, err
	}

	mir, err := syncer.database.Mirror(ctx, act.Mirror)
	if err != nil {
		return nil, nil, nil, err
	}

	cached := new(database.CacheData)

	err = syncer.database.Inode(ctx, cached, mir.Inode)
	if err != nil {
		return nil, nil, nil, err
	}

	return cached, mir, act, nil
}

// activity

type newMirrorActivityArgs struct {
	MirrorID string
	Type     string
	LockCtx  context.Context //nolint:containedctx
	LockConn *sql.Conn
}

func (syncer *syncer) beginActivity(tx database.Transaction, args *newMirrorActivityArgs) (*database.CacheData, *database.Mirror, *database.MirrorActivity, error) {
	var err error
	var ctx context.Context

	if args.LockConn != nil {
		ctx = args.LockCtx
	} else {
		var conn *sql.Conn
		ctx, conn, err = syncer.lock(args.MirrorID, defaultLockWait)
		if err != nil {
			return nil, nil, nil, err
		}
		defer syncer.unlock(args.MirrorID, conn)
	}

	if tx == nil {
		ctx, tx, err = syncer.database.Tx(ctx)
		if err != nil {
			return nil, nil, nil, err
		}
		defer rollback(tx)
	} else {
		ctx = syncer.database.AddTxToCtx(ctx, tx)
	}

	cached, mirror, err := syncer.reverseTraverseToMirror(ctx, args.MirrorID)
	if err != nil {
		return nil, nil, nil, err
	}

	clients := syncer.edb.Clients(ctx)

	unfinishedActivities, err := clients.MirrorActivity.Query().Where(entact.HasMirrorWith(entmir.ID(mirror.ID))).Where(entact.StatusIn(util.MirrorActivityStatusPending, util.MirrorActivityStatusExecuting)).Count(ctx)
	if err != nil {
		return nil, nil, nil, err
	}

	if unfinishedActivities > 0 {
		return nil, nil, nil, errors.New("mirror operations are already underway")
	}

	if !cached.Inode().ReadOnly {
		switch args.Type {
		case util.MirrorActivityTypeLocked:
		default:
			return nil, nil, nil, ErrMirrorLocked
		}
	}

	deadline := time.Now().Add(time.Minute * 20)

	activity, err := syncer.database.CreateMirrorActivity(ctx, &database.CreateMirrorActivityArgs{
		Type:       args.Type,
		Status:     util.MirrorActivityStatusPending,
		EndAt:      time.Now(),
		Mirror:     mirror.ID,
		Namespace:  cached.Namespace.ID,
		Controller: syncer.pubsub.hostname,
		Deadline:   deadline,
	})
	if err != nil {
		return nil, nil, nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, nil, nil, err
	}

	syncer.logToNamespace(ctx, time.Now(), cached, "Commenced new mirror activity '%s' on mirror: %s", args.Type, cached.Path())

	syncer.pubsub.NotifyMirror(cached.Inode())

	syncer.logToMirrorActivity(ctx, time.Now(), cached.Namespace, mirror, activity, "Commenced new mirror activity '%s' on mirror: %s", args.Type, cached.Path())

	// schedule timeouts
	syncer.scheduleTimeout(activity.ID.String(), activity.Controller, deadline)

	return cached, mirror, activity, nil
}

func (syncer *syncer) NewActivity(tx database.Transaction, args *newMirrorActivityArgs) error {
	syncer.sugar.Debugf("Handling mirror activity: %s", this())

	cached, mirror, activity, err := syncer.beginActivity(tx, args)
	if err != nil {
		return err
	}

	go syncer.execute(cached, mirror, activity)

	return nil
}

func (syncer *syncer) execute(cached *database.CacheData, mirror *database.Mirror, activity *database.MirrorActivity) {
	var err error

	defer func() {
		syncer.fail(cached, mirror, activity, err)
	}()

	ctx := context.Background()

	switch activity.Type {
	case util.MirrorActivityTypeInit:
		err = syncer.initMirror(ctx, cached, mirror, activity)
		if err != nil {
			return
		}
	case util.MirrorActivityTypeLocked: // NOTE: intentionally left empty
	case util.MirrorActivityTypeUnlocked: // NOTE: intentionally left empty
	case util.MirrorActivityTypeReconfigure:
		err = syncer.hardSync(ctx, cached, mirror, activity)
		if err != nil {
			return
		}

		// in case of cron
		syncer.server.syncerCronPollerMirror(mirror)

	case util.MirrorActivityTypeCronSync:
		err = syncer.hardSync(ctx, cached, mirror, activity)
		if err != nil {
			return
		}
	case util.MirrorActivityTypeSync:
		err = syncer.hardSync(ctx, cached, mirror, activity)
		if err != nil {
			return
		}
	default:
		syncer.logToMirrorActivity(ctx, time.Now(), cached.Namespace, mirror, activity, "Unrecognized syncer activity type.")
	}

	err = syncer.success(cached, mirror, activity)
	if err != nil {
		return
	}
}

func (syncer *syncer) fail(cached *database.CacheData, mirror *database.Mirror, activity *database.MirrorActivity, e error) {
	if e == nil {
		return
	}

	ctx, conn, err := syncer.lock(mirror.ID.String(), defaultLockWait)
	if err != nil {
		syncer.sugar.Error(err)
		return
	}
	defer syncer.unlock(mirror.ID.String(), conn)

	tctx, tx, err := syncer.database.Tx(ctx)
	if err != nil {
		syncer.sugar.Error(err)
		return
	}
	defer rollback(tx)

	clients := syncer.edb.Clients(tctx)

	act, err := clients.MirrorActivity.Get(ctx, activity.ID)
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

	err = tx.Commit()
	if err != nil {
		syncer.sugar.Error(err)
		return
	}

	syncer.pubsub.NotifyMirror(cached.Inode())

	syncer.logToMirrorActivity(ctx, time.Now(), cached.Namespace, mirror, activity, "Mirror activity '%s' failed: %v", act.Type, e)

	syncer.timers.deleteTimersForActivity(activity.ID.String())
}

func (syncer *syncer) success(cached *database.CacheData, mirror *database.Mirror, activity *database.MirrorActivity) error {
	ctx, conn, err := syncer.lock(mirror.ID.String(), defaultLockWait)
	if err != nil {
		return err
	}
	defer syncer.unlock(mirror.ID.String(), conn)

	tctx, tx, err := syncer.database.Tx(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx)

	clients := syncer.edb.Clients(tctx)

	act, err := clients.MirrorActivity.Get(ctx, activity.ID)
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

	err = tx.Commit()
	if err != nil {
		return err
	}

	syncer.pubsub.NotifyMirror(cached.Inode())

	syncer.logToMirrorActivity(ctx, time.Now(), cached.Namespace, mirror, activity, "Completed mirror activity '%s'.", act.Type)

	syncer.timers.deleteTimersForActivity(activity.ID.String())

	return nil
}

func (syncer *syncer) initMirror(ctx context.Context, cached *database.CacheData, mirror *database.Mirror, activity *database.MirrorActivity) error {
	return syncer.hardSync(ctx, cached, mirror, activity)
}

func (syncer *syncer) tarGzDir(path string) ([]byte, error) {
	tf, err := os.CreateTemp("", "outtar")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tf.Name())

	err = tarGzDir(path, tf)
	if err != nil {
		return nil, err
	}

	return os.ReadFile(tf.Name())
}

func tarGzDir(src string, buf io.Writer) error {
	zr := gzip.NewWriter(buf)
	tw := tar.NewWriter(zr)

	err := filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

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

func (syncer *syncer) hardSync(ctx context.Context, cached *database.CacheData, mirror *database.Mirror, activity *database.MirrorActivity) error {
	lr, err := loadRepository(ctx, &repositorySettings{
		UUID:       mirror.ID,
		URL:        mirror.URL,
		Branch:     mirror.Ref,
		Passphrase: mirror.Passphrase,
		PrivateKey: mirror.PrivateKey,
		PublicKey:  mirror.PublicKey,
	})
	if err != nil {
		return err
	}
	defer lr.Cleanup()

	model, err := buildModel(ctx, lr)
	if err != nil {
		return err
	}

	tctx, tx, err := syncer.database.Tx(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx)

	cached, mirror, err = syncer.reverseTraverseToMirror(tctx, mirror.ID.String())
	if err != nil {
		return err
	}

	cache := make(map[string]*database.Inode)
	trueroot := filepath.Join(cached.Path(), ".")
	if trueroot == "/" {
		cache[trueroot] = cached.Inode()
	} else {
		cache[trueroot+"/"] = cached.Inode()
	}

	var recurser func(parent *database.Inode, path string) error
	recurser = func(parent *database.Inode, path string) error {
		for _, child := range parent.Children {

			rcached := new(database.CacheData)

			err = syncer.database.Inode(tctx, rcached, child.ID)
			if err != nil {
				return err
			}

			child = rcached.Inode()

			cpath := filepath.Join(path, child.Name)

			if child.Type == util.InodeTypeDirectory && child.ExtendedType == util.InodeTypeGit {

				_, err = model.lookup(cpath)
				if errors.Is(err, os.ErrNotExist) {
					continue
				}
				if err != nil {
					return err
				}

				err = syncer.flow.deleteNode(tctx, &deleteNodeArgs{
					cached:    cached,
					super:     true,
					recursive: true,
				})
				if err != nil {
					return err
				}

			} else if child.Type == util.InodeTypeDirectory {

				err = recurser((child), cpath)
				if err != nil {
					return err
				}

				mn, err := model.lookup(cpath)
				if err != nil && !errors.Is(err, os.ErrNotExist) {
					return err
				}

				if errors.Is(err, os.ErrNotExist) || mn.ntype != mntDir {

					err = syncer.flow.deleteNode(tctx, &deleteNodeArgs{
						cached:    cached,
						super:     true,
						recursive: true,
					})
					if err != nil {
						return err
					}
				}

			} else if child.Type == util.InodeTypeWorkflow {

				mn, err := model.lookup(cpath)
				if err != nil && !errors.Is(err, os.ErrNotExist) {
					return err
				}

				if errors.Is(err, os.ErrNotExist) || mn.ntype != mntWorkflow {

					err = syncer.flow.deleteNode(tctx, &deleteNodeArgs{
						cached:    cached,
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

	err = recurser(cached.Inode(), ".")
	if err != nil {
		return err
	}

	err = modelWalk(model.root, ".", func(path string, n *mirrorNode, err error) error {
		if path == "." {
			return nil
		}

		truepath := filepath.Join(cached.Path(), path)
		dir, _ := filepath.Split(truepath)

		switch n.ntype {
		case mntDir:

			pino := cache[dir]
			pcached := new(database.CacheData)
			err := syncer.database.Inode(tctx, pcached, pino.ID)
			if err != nil {
				return err
			}

			ino, err := syncer.flow.createDirectory(tctx, &createDirectoryArgs{
				pcached: pcached,
				path:    truepath,
				super:   true,
			})
			if ino == nil {
				return err
			}

			cache[truepath+"/"] = ino

		case mntWorkflow:

			data, err := os.ReadFile(filepath.Join(lr.path, path+n.extension))
			if err != nil {
				return err
			}

			wf, ino, err := syncer.flow.createWorkflow(tctx, &createWorkflowArgs{
				ns:         cached.Namespace,
				pino:       cache[dir],
				path:       truepath,
				super:      true,
				data:       data,
				noValidate: true,
			})
			if wf == nil {
				return err
			}
			if errors.Is(err, os.ErrExist) {

				ucached := new(database.CacheData)
				err = syncer.database.Workflow(tctx, ucached, wf.ID)
				if err != nil {
					return err
				}

				_, err = syncer.flow.updateWorkflow(tctx, &updateWorkflowArgs{
					cached:     ucached,
					path:       truepath,
					super:      true,
					data:       data,
					noValidate: true,
				})
				if err != nil {
					return err
				}

				cache[truepath] = ucached.Inode()
			} else {
				cache[truepath] = ino
			}

		case mntNamespaceVar:

			var data []byte
			var err error

			fpath := filepath.Join(lr.path, path)

			if n.isDir {
				data, err = syncer.tarGzDir(fpath)
			} else {
				data, err = os.ReadFile(fpath)
			}
			if err != nil {
				return err
			}

			_, base := filepath.Split(path)
			trimmed := strings.TrimPrefix(base, "var.")

			_, _, err = syncer.flow.SetVariable(tctx, &entNamespaceVarQuerier{cached: cached, clients: syncer.edb.Clients(tctx)}, trimmed, data, "", false)
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
				data, err = os.ReadFile(fpath)
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
			_, base := filepath.Split(x[0])

			child := cache[dir+base]

			if child == nil || child.Type != util.InodeTypeWorkflow {
				if derrors.IsNotFound(err) {
					syncer.logToMirrorActivity(tctx, time.Now(), cached.Namespace, mirror, activity, "Found something that looks like a workflow variable with no matching workflow: "+cached.Path())
					return nil
				}
			}

			wcached, err := syncer.flow.reverseTraverseToWorkflow(tctx, child.Workflow.String())
			if err != nil {
				return err
			}

			_, _, err = syncer.flow.SetVariable(tctx, &entWorkflowVarQuerier{cached: wcached, clients: syncer.edb.Clients(tctx)}, trimmed, data, "", false)
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
		return err
	}

	syncer.flow.database.InvalidateNamespace(ctx, cached, true)

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
	_ = os.RemoveAll(repository.path)
}

func (repository *localRepository) clone(ctx context.Context) error {
	uri := repository.repo.URL
	prefix := "https://"

	cloneOptions := &git.CloneOptions{
		URL:           uri,
		Progress:      os.Stdout,
		ReferenceName: plumbing.NewBranchReferenceName(repository.repo.Branch),
	}

	// https with access token case. Put passphrase inside the git url.
	if strings.HasPrefix(uri, prefix) && len(repository.repo.Passphrase) > 0 {
		if !strings.Contains(uri, "@") {
			uri = fmt.Sprintf("%s%s@", prefix, repository.repo.Passphrase) + strings.TrimPrefix(uri, prefix)
		}
	}

	// ssh case. Configure cloneOptions.Auth field.
	if !strings.HasPrefix(uri, prefix) {
		publicKeys, err := gitSSH.NewPublicKeys("git", []byte(repository.repo.PrivateKey), repository.repo.Passphrase)
		if err != nil {
			return err
		}
		publicKeys.HostKeyCallbackHelper = gitSSH.HostKeyCallbackHelper{
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
		cloneOptions.Auth = publicKeys
	}

	r, err := git.PlainClone(repository.path, false, cloneOptions)
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

type mirrorNode struct {
	parent    *mirrorNode
	children  []*mirrorNode
	ntype     string
	name      string
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

// var (
// 	ntypeShort map[string]string
// )

// func init() {
// 	ntypeShort = map[string]string{
// 		mntDir:      "d",
// 		mntWorkflow: "w",
// 	}
// }

// func (model *mirrorModel) dump() {
// 	err := modelWalk(model.root, ".", func(path string, n *mirrorNode, err error) error {
// 		fmt.Println(ntypeShort[n.ntype], path, n.change)
// 		return nil
// 	})
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// }

func modelWalk(node *mirrorNode, path string, fn func(path string, n *mirrorNode, err error) error) error {
	err := fn(path, node, nil)
	if errors.Is(err, filepath.SkipDir) {
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

func buildModel(ctx context.Context, repo *localRepository) (*mirrorModel, error) {
	model := new(mirrorModel)
	model.root = &mirrorNode{
		parent:   model.root,
		children: make([]*mirrorNode, 0),
		ntype:    mntDir,
		name:     ".",
	}

	cfg := new(project.Config)

	scfpath := filepath.Join(repo.path, project.ConfigFile)
	scfgbytes, err := os.ReadFile(scfpath)
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
		if err != nil {
			return err
		}

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

	// model.dump()

	return model, nil
}
