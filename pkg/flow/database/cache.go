package database

import (
	"context"
	"encoding/json"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	gocache_store "github.com/eko/gocache/store/go_cache/v4"
	"github.com/google/uuid"
	gocache "github.com/patrickmn/go-cache"
	"go.uber.org/zap"
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

	if !cacheHit {
		db.storeInstanceInCache(ctx, inst)
	}

	if cached.Namespace == nil {
		err = db.Namespace(ctx, cached, cached.Instance.Namespace)
		if err != nil {
			return err
		}
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
