package database

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	gocache_store "github.com/eko/gocache/store/go_cache/v4"
	"github.com/google/uuid"
	gocache "github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

type CachedDatabase struct {
	sugar  *zap.SugaredLogger
	source Database
	cache  *cache.Cache[[]byte]
}

func NewCachedDatabase(sugar *zap.SugaredLogger, source Database) *CachedDatabase {
	db := &CachedDatabase{
		sugar:  sugar,
		source: source,
	}
	gocacheClient := gocache.New(5*time.Minute, 10*time.Minute)
	gocacheStore := gocache_store.NewGoCache(gocacheClient)
	db.cache = cache.New[[]byte](gocacheStore)
	db.sugar.Warnf("Initializing cache.")
	return db
}

func (db *CachedDatabase) Close() error {
	return db.source.Close()
}

func (db *CachedDatabase) Tx(ctx context.Context) (Transaction, error) {
	return db.source.Tx(ctx)
}

func (db *CachedDatabase) Namespace(ctx context.Context, tx Transaction, cached *CacheData, id uuid.UUID) error {

	if tx == nil {

		ns := db.lookupNamespaceByID(ctx, id)

		if ns != nil {
			cached.Namespace = ns
			return nil
		}
	}

	ns, err := db.source.Namespace(ctx, tx, id)
	if err != nil {
		return err
	}

	cached.Namespace = ns

	if tx == nil {
		db.storeNamespaceInCache(ctx, cached.Namespace)
	}

	return nil

}

func (db *CachedDatabase) NamespaceByName(ctx context.Context, tx Transaction, cached *CacheData, name string) error {

	if tx == nil {

		ns := db.lookupNamespaceByName(ctx, name)

		if ns != nil {
			cached.Namespace = ns
			return nil
		}
	}

	ns, err := db.source.NamespaceByName(ctx, tx, name)
	if err != nil {
		return err
	}

	cached.Namespace = ns

	if tx == nil {
		db.storeNamespaceInCache(ctx, cached.Namespace)
	}

	return nil

}

func (db *CachedDatabase) Inode(ctx context.Context, tx Transaction, cached *CacheData, id uuid.UUID) error {

	var err error
	var ino *Inode

	if tx == nil {
		ino = db.lookupInodeByID(ctx, id)
	}

	if ino == nil {
		ino, err = db.source.Inode(ctx, tx, id)
		if err != nil {
			return err
		}
	}

	if ino.Name != "" {
		err = db.Inode(ctx, tx, cached, ino.Parent)
		if err != nil {
			return err
		}

		cached.Inodes = append(cached.Inodes, ino)
	}

	if cached.Namespace == nil {
		err = db.Namespace(ctx, tx, cached, ino.Namespace)
		if err != nil {
			return err
		}
	}

	if tx == nil {
		db.storeInodeInCache(ctx, ino)
	}

	return nil

}

func (db *CachedDatabase) InodeByPath(ctx context.Context, tx Transaction, cached *CacheData, path string) error {

	if cached.Namespace == nil {
		panic("this function should not be called unless the namespace has already been resolved")
	}

	path = filepath.Join("/", path)
	elems := filepath.SplitList(path)

	err := db.Inode(ctx, tx, cached, cached.Namespace.Root)
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

		err = db.Inode(ctx, tx, cached, ino.ID)
		if err != nil {
			return err
		}
	}

	return nil

}

func (db *CachedDatabase) Workflow(ctx context.Context, tx Transaction, cached *CacheData, id uuid.UUID) error {

	var err error
	var wf *Workflow

	if tx == nil {
		wf = db.lookupWorkflowByID(ctx, id)
	}

	if wf == nil {
		wf, err = db.source.Workflow(ctx, tx, id)
		if err != nil {
			return err
		}
	}

	cached.Workflow = wf

	if cached.Inodes == nil {
		err = db.Inode(ctx, tx, cached, wf.Inode)
		if err != nil {
			return err
		}
	}

	if tx == nil {
		db.storeWorkflowInCache(ctx, wf)
	}

	return nil

}

func (db *CachedDatabase) Revision(ctx context.Context, tx Transaction, cached *CacheData, id uuid.UUID) error {

	var err error
	var rev *Revision

	if tx == nil {
		rev = db.lookupRevisionByID(ctx, id)
	}

	if rev == nil {
		rev, err = db.source.Revision(ctx, tx, id)
		if err != nil {
			return err
		}
	}

	cached.Revision = rev

	if tx == nil {
		db.storeRevisionInCache(ctx, rev)
	}

	return nil

}

func (db *CachedDatabase) Instance(ctx context.Context, tx Transaction, cached *CacheData, id uuid.UUID) error {

	var err error
	var inst *Instance

	if tx == nil {
		inst = db.lookupInstanceByID(ctx, id)
	}

	inst, err = db.source.Instance(ctx, tx, id)
	if err != nil {
		return err
	}

	cached.Instance = inst

	if cached.Workflow == nil {
		err = db.Workflow(ctx, tx, cached, cached.Instance.Workflow)
		if err != nil {
			return err
		}
	}

	if tx == nil {
		db.storeInstanceInCache(ctx, inst)
	}

	return nil

}

func (db *CachedDatabase) InstanceRuntime(ctx context.Context, tx Transaction, id uuid.UUID) (*InstanceRuntime, error) {
	// NOTE: not bothering to cache this right now
	return db.source.InstanceRuntime(ctx, tx, id)
}

func (db *CachedDatabase) NamespaceAnnotation(ctx context.Context, tx Transaction, inodeID uuid.UUID, key string) (*Annotation, error) {
	// NOTE: not bothering to cache this right now
	return db.source.NamespaceAnnotation(ctx, tx, inodeID, key)
}

func (db *CachedDatabase) InodeAnnotation(ctx context.Context, tx Transaction, inodeID uuid.UUID, key string) (*Annotation, error) {
	// NOTE: not bothering to cache this right now
	return db.source.InodeAnnotation(ctx, tx, inodeID, key)
}

func (db *CachedDatabase) WorkflowAnnotation(ctx context.Context, tx Transaction, wfID uuid.UUID, key string) (*Annotation, error) {
	// NOTE: not bothering to cache this right now
	return db.source.WorkflowAnnotation(ctx, tx, wfID, key)
}

func (db *CachedDatabase) InstanceAnnotation(ctx context.Context, tx Transaction, instID uuid.UUID, key string) (*Annotation, error) {
	// NOTE: not bothering to cache this right now
	return db.source.InstanceAnnotation(ctx, tx, instID, key)
}

func (db *CachedDatabase) ThreadVariables(ctx context.Context, tx Transaction, instID uuid.UUID) ([]*VarRef, error) {
	// NOTE: not bothering to cache this right now
	return db.source.ThreadVariables(ctx, tx, instID)
}

func (db *CachedDatabase) NamespaceVariable(ctx context.Context, tx Transaction, nsID uuid.UUID, key string) (*VarRef, error) {
	// NOTE: not bothering to cache this right now
	return db.source.InstanceVariableRef(ctx, tx, nsID, key)
}

func (db *CachedDatabase) WorkflowVariable(ctx context.Context, tx Transaction, wfID uuid.UUID, key string) (*VarRef, error) {
	// NOTE: not bothering to cache this right now
	return db.source.WorkflowVariableRef(ctx, tx, wfID, key)
}

func (db *CachedDatabase) InstanceVariable(ctx context.Context, tx Transaction, instID uuid.UUID, key string) (*VarRef, error) {
	// NOTE: not bothering to cache this right now
	return db.source.InstanceVariableRef(ctx, tx, instID, key)
}

func (db *CachedDatabase) ThreadVariable(ctx context.Context, tx Transaction, instID uuid.UUID, key string) (*VarRef, error) {
	// NOTE: not bothering to cache this right now
	return db.source.ThreadVariableRef(ctx, tx, instID, key)
}

func (db *CachedDatabase) VariableData(ctx context.Context, tx Transaction, id uuid.UUID, load bool) (*VarData, error) {
	// NOTE: not bothering to cache this right now
	return db.source.VariableData(ctx, tx, id, load)
}

func (db *CachedDatabase) Mirror(ctx context.Context, tx Transaction, id uuid.UUID) (*Mirror, error) {
	// NOTE: not bothering to cache this right now
	return db.source.Mirror(ctx, tx, id)
}
