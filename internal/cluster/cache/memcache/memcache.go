package memcache

import (
	"context"
	"log/slog"
	"time"
	"unsafe"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/direktiv/direktiv/internal/cluster/cache"
	"github.com/direktiv/direktiv/internal/cluster/pubsub"
	"github.com/direktiv/direktiv/internal/core"
)

const (
	defaultCacheTTL = 5 * time.Minute
)

// type cacheMessage struct {
// 	Hostname string
// 	Key      string
// }

type CacheManager struct {
	bus pubsub.EventBus

	cacheSecret *Cache[core.Secret]
	cacheFlow   *Cache[core.TypescriptFlow]
}

type Cache[T any] struct {
	name  string
	cache *ristretto.Cache[string, T]
	bus   pubsub.EventBus
}

func NewManager(bus pubsub.EventBus) (*CacheManager, error) {
	secrets, err := newCache[core.Secret](core.SecretsCacheName, bus)
	if err != nil {
		slog.Error("cannot create secret cache", slog.Any("error", err))
		return nil, err
	}

	flows, err := newCache[core.TypescriptFlow](core.FlowCacheName, bus)
	if err != nil {
		slog.Error("cannot create flow cache", slog.Any("error", err))
		return nil, err
	}

	return &CacheManager{
		cacheSecret: secrets,
		cacheFlow:   flows,
		bus:         bus,
	}, nil
}

func newCache[T any](name string, bus pubsub.EventBus) (*Cache[T], error) {
	c, err := ristretto.NewCache(&ristretto.Config[string, T]{
		NumCounters: 10000000,
		MaxCost:     1 << 30,
		BufferItems: 64,
		OnEvict: func(item *ristretto.Item[T]) {
			slog.Info("cache item evicted", slog.String("cache", name),
				slog.Uint64("key", item.Key))
		},
	})
	if err != nil {
		slog.Error("cannot create cache", slog.String("name", name),
			slog.Any("error", err))
		return nil, err
	}

	return &Cache[T]{
		name:  name,
		cache: c,
		bus:   bus,
	}, nil
}

func (cm *CacheManager) SecretsCache() cache.Cache[core.Secret] {
	return cm.cacheSecret
}

func (cm *CacheManager) FlowCache() cache.Cache[core.TypescriptFlow] {
	return cm.cacheFlow
}

func (c *Cache[T]) Get(key string, fetch func(...any) (T, error)) (T, error) {
	var storeValue T
	var err error

	slog.Info("get key from cache", slog.String("cache", c.name), slog.String("key", key))

	v, found := c.cache.Get(key)
	if found {
		slog.Info("found item in cache", slog.String("key", key))

		return v, nil
	}

	storeValue, err = fetch()
	if err != nil {
		return storeValue, err
	}

	c.Set(key, storeValue)

	return storeValue, err
}

func (c *Cache[T]) Delete(key string) {
	slog.Info("delete item from cache", slog.String("key", key))
	c.cache.Del(key)
}

func (c *Cache[T]) Set(key string, value T) {
	slog.Info("set key in cache", slog.String("cache", c.name), slog.String("key", key))

	c.cache.SetWithTTL(key, value,
		int64(unsafe.Sizeof(value)), defaultCacheTTL)
}

func (c *Cache[T]) Notify(ctx context.Context, notify cache.CacheNotify) {
	c.Delete(notify.Key)
}
