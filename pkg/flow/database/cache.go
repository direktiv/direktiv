package database

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/util"
	"github.com/eko/gocache/lib/v4/cache"
	gocache_store "github.com/eko/gocache/store/go_cache/v4"
	"github.com/google/uuid"
	gocache "github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	PubsubNotifyFunction = "cache"
)

type Notifier interface {
	PublishToCluster(string)
}

type notification struct {
	Operation string
	ID        uuid.UUID
	Recursive bool
}

func (n *notification) Marshal() string {
	data, err := json.Marshal(n)
	if err != nil {
		panic(err)
	}
	return string(data)
}

type CachedDatabase struct {
	sugar          *zap.SugaredLogger
	source         Database
	cache          *cache.Cache[[]byte]
	notifier       Notifier
	cachingEnabled bool
}

func NewCachedDatabase(sugar *zap.SugaredLogger, source Database, notifier Notifier) *CachedDatabase {
	db := &CachedDatabase{
		sugar:    sugar,
		source:   source,
		notifier: notifier,
	}
	gocacheClient := gocache.New(5*time.Minute, 10*time.Minute)
	gocacheStore := gocache_store.NewGoCache(gocacheClient)
	db.cache = cache.New[[]byte](gocacheStore)
	db.sugar.Warnf("Initializing cache.")
	return db
}

func (db *CachedDatabase) HandleNotification(s string) {
	notification := new(notification)

	err := json.Unmarshal([]byte(s), &notification)
	if err != nil {
		db.sugar.Error(err)
		return
	}

	switch notification.Operation {
	case "invalidate-namespace":
		db.invalidateCachedNamespace(context.Background(), notification.ID, notification.Recursive)
	case "invalidate-inode":
		db.invalidateCachedNamespace(context.Background(), notification.ID, notification.Recursive)
	case "invalidate-workflow":
		db.invalidateCachedWorkflow(context.Background(), notification.ID, notification.Recursive)
	default:
		db.sugar.Error(err)
		return
	}
}

func (db *CachedDatabase) Close() error {
	return db.source.Close()
}

func (db *CachedDatabase) AddTxToCtx(ctx context.Context, tx Transaction) context.Context {
	return db.source.AddTxToCtx(ctx, tx)
}

func (db *CachedDatabase) Tx(ctx context.Context) (context.Context, Transaction, error) {
	return db.source.Tx(ctx)
}

func (db *CachedDatabase) Namespace(ctx context.Context, cached *CacheData, id uuid.UUID) error {
	ns := db.lookupNamespaceByID(ctx, id)

	if ns != nil {
		cached.Namespace = ns
		return nil
	}

	ns, err := db.source.Namespace(ctx, id)
	if err != nil {
		return err
	}

	cached.Namespace = ns

	db.storeNamespaceInCache(ctx, cached.Namespace)

	return nil
}

func (db *CachedDatabase) NamespaceByName(ctx context.Context, cached *CacheData, name string) error {
	var err error

	ns := db.lookupNamespaceByName(ctx, name)

	if ns != nil {
		cached.Namespace = ns
		return nil
	} else {
		ns, err = db.source.NamespaceByName(ctx, name)
		if err != nil {
			return err
		}
	}

	cached.Namespace = ns

	db.storeNamespaceInCache(ctx, cached.Namespace)

	return nil
}

func (db *CachedDatabase) InvalidateNamespace(ctx context.Context, cached *CacheData, recursive bool) {
	db.notifier.PublishToCluster((&notification{
		Operation: "invalidate-namespace",
		ID:        cached.Namespace.ID,
		Recursive: recursive,
	}).Marshal())

	db.invalidateCachedNamespace(ctx, cached.Namespace.ID, recursive)
}

func (db *CachedDatabase) Inode(ctx context.Context, cached *CacheData, id uuid.UUID) error {
	var err error

	cacheHit := true

	ino := db.lookupInodeByID(ctx, id)

	if ino == nil {
		cacheHit = false
		ino, err = db.source.Inode(ctx, id)
		if err != nil {
			return err
		}
	}

	if ino.Name != "" {
		err = db.Inode(ctx, cached, ino.Parent)
		if err != nil {
			return err
		}
		cached.Inodes = append(cached.Inodes, ino)
	} else {
		cached.Inodes = []*Inode{ino}
	}

	if cached.Namespace == nil {
		err = db.Namespace(ctx, cached, ino.Namespace)
		if err != nil {
			return err
		}
	}

	if !cacheHit {
		db.storeInodeInCache(ctx, ino)
	}

	return nil
}

func (db *CachedDatabase) InodeByPath(ctx context.Context, cached *CacheData, path string) error {
	if cached.Namespace == nil {
		panic("this function should not be called unless the namespace has already been resolved")
	}

	path = filepath.Join("/", path)
	if path == "/" {
		path = ""
	}

	elems := strings.Split(path, "/")

	err := db.Inode(ctx, cached, cached.Namespace.Root)
	if err != nil {
		return err
	}

	if len(elems) < 2 {
		return nil
	}

	for i := 1; i < len(elems); i++ {

		pino := cached.Inodes[i-1]
		name := elems[i]

		var ino *Inode

		for j := range pino.Children {
			x := pino.Children[j]
			if x.Name == name {
				ino = x
				break
			}
		}

		if ino == nil {
			return os.ErrNotExist
		}

		err = db.Inode(ctx, cached, ino.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *CachedDatabase) InvalidateInode(ctx context.Context, cached *CacheData, recursive bool) {
	db.notifier.PublishToCluster((&notification{
		Operation: "invalidate-inode",
		ID:        cached.Inode().ID,
		Recursive: recursive,
	}).Marshal())

	db.invalidateCachedInode(ctx, cached.Inode().ID, recursive)
}

func (db *CachedDatabase) CreateDirectoryInode(ctx context.Context, args *CreateDirectoryInodeArgs) (*Inode, error) {
	if args.Parent.Type != util.InodeTypeDirectory {
		return nil, status.Error(codes.AlreadyExists, "parent node is not a directory")
	}

	for i := range args.Parent.Children {
		child := args.Parent.Children[i]
		if child.Name == args.Name {
			if child.Type == util.InodeTypeDirectory {
				cached := new(CacheData)
				err := db.Inode(ctx, cached, child.ID)
				if err != nil {
					return nil, err
				}
				return cached.Inode(), os.ErrExist
			}
			return nil, os.ErrExist
		}
	}

	ino, err := db.source.CreateInode(ctx, &CreateInodeArgs{
		Name:      args.Name,
		Type:      util.InodeTypeDirectory,
		ReadOnly:  args.ReadOnly,
		Namespace: args.Parent.Namespace,
		Parent:    args.Parent.ID,
	})
	if err != nil {
		return nil, err
	}

	args.Parent.addChild(ino)

	pino, err := db.source.UpdateInode(ctx, &UpdateInodeArgs{
		Inode: args.Parent,
	})
	if err != nil {
		return nil, err
	}

	*args.Parent = *pino

	// TODO: add to cache and cache invalidate anything relevant

	return ino, nil
}

func (db *CachedDatabase) UpdateInode(ctx context.Context, args *UpdateInodeArgs) (*Inode, error) {
	// TODO: add to cache and cache invalidate anything relevant
	return db.source.UpdateInode(ctx, args)
}

func (db *CachedDatabase) Workflow(ctx context.Context, cached *CacheData, id uuid.UUID) error {
	var err error

	cacheHit := true

	wf := db.lookupWorkflowByID(ctx, id)

	if wf == nil {
		cacheHit = false
		wf, err = db.source.Workflow(ctx, id)
		if err != nil {
			return err
		}
	}

	cached.Workflow = wf

	if cached.Inodes == nil {
		err = db.Inode(ctx, cached, wf.Inode)
		if err != nil {
			return err
		}
	}

	if !cacheHit {
		db.storeWorkflowInCache(ctx, wf)
	}

	return nil
}

func (db *CachedDatabase) InvalidateWorkflow(ctx context.Context, cached *CacheData, recursive bool) {
	db.notifier.PublishToCluster((&notification{
		Operation: "invalidate-workflow",
		ID:        cached.Workflow.ID,
		Recursive: recursive,
	}).Marshal())

	db.invalidateCachedWorkflow(ctx, cached.Workflow.ID, recursive)
}

func (db *CachedDatabase) CreateCompleteWorkflow(ctx context.Context, args *CreateCompleteWorkflowArgs) (*CacheData, error) {
	if args.Parent.Inode().Type != util.InodeTypeDirectory {
		return nil, status.Error(codes.AlreadyExists, "parent node is not a directory")
	}

	for i := range args.Parent.Inode().Children {
		child := args.Parent.Inode().Children[i]
		if child.Name == args.Name {
			return nil, os.ErrExist
		}
	}

	ino, err := db.source.CreateInode(ctx, &CreateInodeArgs{
		Name:      args.Name,
		Type:      util.InodeTypeWorkflow,
		ReadOnly:  args.ReadOnly,
		Namespace: args.Parent.Inode().Namespace,
		Parent:    args.Parent.Inode().ID,
	})
	if err != nil {
		return nil, err
	}

	args.Parent.Inode().addChild(ino)

	pino, err := db.source.UpdateInode(ctx, &UpdateInodeArgs{
		Inode: args.Parent.Inode(),
	})
	if err != nil {
		return nil, err
	}

	*args.Parent.Inode() = *pino

	cached := new(CacheData)
	*cached = *args.Parent
	cached.Inodes = make([]*Inode, 0)
	copy(cached.Inodes, args.Parent.Inodes)
	cached.Inodes = append(cached.Inodes, ino)

	wf, err := db.source.CreateWorkflow(ctx, &CreateWorkflowArgs{
		Inode: ino,
	})
	if err != nil {
		return nil, err
	}

	ino.Workflow = wf.ID
	cached.Workflow = wf

	rev, err := db.source.CreateRevision(ctx, &CreateRevisionArgs{
		Hash:     args.Hash,
		Source:   args.Source,
		Metadata: args.Metadata,
		Workflow: wf.ID,
	})
	if err != nil {
		return nil, err
	}

	cached.Revision = rev

	ref, err := db.source.CreateRef(ctx, &CreateRefArgs{
		Immutable: false,
		Name:      "latest",
		Workflow:  wf.ID,
		Revision:  rev.ID,
	})
	if err != nil {
		return nil, err
	}

	cached.Ref = ref

	cached.Workflow.Refs = []*Ref{ref}
	cached.Workflow.Revisions = []*Revision{rev}

	// TODO: add to cache and cache invalidate anything relevant

	return cached, nil
}

func (db *CachedDatabase) UpdateWorkflow(ctx context.Context, args *UpdateWorkflowArgs) (*Workflow, error) {
	// TODO: add to cache and cache invalidate anything relevant
	return db.source.UpdateWorkflow(ctx, args)
}

func (db *CachedDatabase) Revision(ctx context.Context, cached *CacheData, id uuid.UUID) error {
	var err error

	cacheHit := true

	rev := db.lookupRevisionByID(ctx, id)

	if rev == nil {
		cacheHit = false
		rev, err = db.source.Revision(ctx, id)
		if err != nil {
			return err
		}
	}

	cached.Revision = rev

	if cached.Workflow == nil {
		err = db.Workflow(ctx, cached, cached.Revision.Workflow)
		if err != nil {
			return err
		}
	}

	if !cacheHit {
		db.storeRevisionInCache(ctx, rev)
	}

	return nil
}

func (db *CachedDatabase) CreateRevision(ctx context.Context, args *CreateRevisionArgs) (*Revision, error) {
	// TODO: add to cache and cache invalidate anything relevant
	return db.source.CreateRevision(ctx, args)
}

func (db *CachedDatabase) Instance(ctx context.Context, cached *CacheData, id uuid.UUID) error {
	var err error

	cacheHit := true

	inst := db.lookupInstanceByID(ctx, id)

	if inst == nil {
		cacheHit = false
		inst, err = db.source.Instance(ctx, id)
		if err != nil {
			return err
		}
	}

	cached.Instance = inst

	if cached.Revision == nil {
		err = db.Revision(ctx, cached, cached.Instance.Revision)
		if err != nil {
			return err
		}
	}

	if cached.Workflow == nil {
		err = db.Workflow(ctx, cached, cached.Instance.Workflow)
		if err != nil {
			return err
		}
	}

	if !cacheHit {
		db.storeInstanceInCache(ctx, inst)
	}

	return nil
}

func (db *CachedDatabase) FlushInstance(ctx context.Context, inst *Instance) error {
	db.storeInstanceInCache(ctx, inst)
	return nil
}

func (db *CachedDatabase) InstanceRuntime(ctx context.Context, id uuid.UUID) (*InstanceRuntime, error) {
	// NOTE: not bothering to cache this right now
	return db.source.InstanceRuntime(ctx, id)
}

func (db *CachedDatabase) NamespaceAnnotation(ctx context.Context, inodeID uuid.UUID, key string) (*Annotation, error) {
	// NOTE: not bothering to cache this right now
	return db.source.NamespaceAnnotation(ctx, inodeID, key)
}

func (db *CachedDatabase) InodeAnnotation(ctx context.Context, inodeID uuid.UUID, key string) (*Annotation, error) {
	// NOTE: not bothering to cache this right now
	return db.source.InodeAnnotation(ctx, inodeID, key)
}

func (db *CachedDatabase) WorkflowAnnotation(ctx context.Context, wfID uuid.UUID, key string) (*Annotation, error) {
	// NOTE: not bothering to cache this right now
	return db.source.WorkflowAnnotation(ctx, wfID, key)
}

func (db *CachedDatabase) InstanceAnnotation(ctx context.Context, instID uuid.UUID, key string) (*Annotation, error) {
	// NOTE: not bothering to cache this right now
	return db.source.InstanceAnnotation(ctx, instID, key)
}

func (db *CachedDatabase) ThreadVariables(ctx context.Context, instID uuid.UUID) ([]*VarRef, error) {
	// NOTE: not bothering to cache this right now
	return db.source.ThreadVariables(ctx, instID)
}

func (db *CachedDatabase) NamespaceVariable(ctx context.Context, nsID uuid.UUID, key string) (*VarRef, error) {
	// NOTE: not bothering to cache this right now
	return db.source.NamespaceVariableRef(ctx, nsID, key)
}

func (db *CachedDatabase) WorkflowVariable(ctx context.Context, wfID uuid.UUID, key string) (*VarRef, error) {
	// NOTE: not bothering to cache this right now
	return db.source.WorkflowVariableRef(ctx, wfID, key)
}

func (db *CachedDatabase) InstanceVariable(ctx context.Context, instID uuid.UUID, key string) (*VarRef, error) {
	// NOTE: not bothering to cache this right now
	return db.source.InstanceVariableRef(ctx, instID, key)
}

func (db *CachedDatabase) ThreadVariable(ctx context.Context, instID uuid.UUID, key string) (*VarRef, error) {
	// NOTE: not bothering to cache this right now
	return db.source.ThreadVariableRef(ctx, instID, key)
}

func (db *CachedDatabase) VariableData(ctx context.Context, id uuid.UUID, load bool) (*VarData, error) {
	// NOTE: not bothering to cache this right now
	return db.source.VariableData(ctx, id, load)
}

func (db *CachedDatabase) Mirror(ctx context.Context, id uuid.UUID) (*Mirror, error) {
	// NOTE: not bothering to cache this right now
	return db.source.Mirror(ctx, id)
}

func (db *CachedDatabase) Mirrors(ctx context.Context) ([]uuid.UUID, error) {
	// NOTE: not bothering to cache this right now
	return db.source.Mirrors(ctx)
}

func (db *CachedDatabase) MirrorActivity(ctx context.Context, id uuid.UUID) (*MirrorActivity, error) {
	// NOTE: not bothering to cache this right now
	return db.source.MirrorActivity(ctx, id)
}

func (db *CachedDatabase) CreateMirrorActivity(ctx context.Context, args *CreateMirrorActivityArgs) (*MirrorActivity, error) {
	// NOTE: not bothering to cache this right now
	return db.source.CreateMirrorActivity(ctx, args)
}
