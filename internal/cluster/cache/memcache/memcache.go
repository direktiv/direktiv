package memcache

import (
	"context"
	"encoding/json"
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

	cacheSecret     *Cache[core.Secret]
	cacheFlow       *Cache[core.TypescriptFlow]
	cacheNamespaces *Cache[[]string]
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

	namespaces, err := newCache[[]string]("namespaces", bus)
	if err != nil {
		slog.Error("cannot create namespace cache", slog.Any("error", err))
		return nil, err
	}

	return &CacheManager{
		cacheSecret:     secrets,
		cacheFlow:       flows,
		cacheNamespaces: namespaces,
		bus:             bus,
	}, nil
}

func newCache[T any](name string, bus pubsub.EventBus) (*Cache[T], error) {
	c, err := ristretto.NewCache(&ristretto.Config[string, T]{
		NumCounters: 10000000,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}

	cacheInstance := &Cache[T]{
		name:  name,
		cache: c,
		bus:   bus,
	}

	// SUBSCRIBE to cache invalidation events
	err = bus.Subscribe(pubsub.SubjCacheDelete, func(data []byte) {
		var notify cache.CacheNotify
		if err := json.Unmarshal(data, &notify); err != nil {
			slog.Error("failed to unmarshal cache notify", slog.Any("error", err))
			return
		}
		// IMPORTANT: apply locally ONLY
		cacheInstance.applyNotify(notify)
	})
	if err != nil {
		return nil, err
	}

	return cacheInstance, nil
}

func (cm *CacheManager) NamespaceCache() cache.Cache[[]string] {
	return cm.cacheNamespaces
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
	// 1. Apply locally
	c.applyNotify(notify)

	// 2. Broadcast to cluster
	b, err := json.Marshal(notify)
	if err != nil {
		return
	}

	err = c.bus.Publish(pubsub.SubjCacheDelete, b)
	if err != nil {
		slog.Error("failed to publish cache delete event", slog.Any("error", err))

		return
	}
}

func (c *Cache[T]) applyNotify(notify cache.CacheNotify) {
	slog.Info("apply cache notify",
		slog.String("cache", c.name),
		slog.String("key", notify.Key),
		slog.String("action", string(notify.Action)),
	)
	switch notify.Action {
	case cache.CacheClear:
		c.cache.Clear()
	default:
		c.cache.Del(notify.Key)
	}
}
