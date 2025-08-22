package cache

import (
	"log/slog"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/pubsub"
)

type Cache struct {
	cache *ristretto.Cache[string, []byte]
	bus   *pubsub.Bus
}

func NewCache(bus *pubsub.Bus, enableMetrics bool) (*Cache, error) {
	cache, err := ristretto.NewCache(&ristretto.Config[string, []byte]{
		NumCounters: 10000000,
		MaxCost:     1073741824,
		BufferItems: 64,
		Metrics:     enableMetrics,
	})
	if err != nil {
		return nil, err
	}

	return &Cache{
		bus:   bus,
		cache: cache,
	}, nil
}

func (c *Cache) SetTTL(key string, value []byte, ttl int) {
	if ttl == 0 {
		c.cache.Set(key, value, int64(len(value)))
	} else {
		c.cache.SetWithTTL(key, value, int64(len(value)),
			time.Duration(ttl)*time.Second)
	}

	c.cache.Wait()
	err := c.bus.Publish(pubsub.CacheDeleteEvent, []byte(key))
	if err != nil {
		slog.Error("can not publish cache invalidate", slog.Any("error", err))
	}
}

func (c *Cache) Set(key string, value []byte) {
	c.SetTTL(key, value, 0)
}

func (c *Cache) Delete(key string) {
	c.cache.Del(key)
	err := c.bus.Publish(pubsub.CacheDeleteEvent, []byte(key))
	if err != nil {
		slog.Error("can not publish cache invalidate", slog.Any("error", err))
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	return c.cache.Get(key)
}

func (c *Cache) Run(circuit *core.Circuit) {
	c.bus.Subscribe(pubsub.CacheDeleteEvent, func(data []byte) {
		c.cache.Del(string(data))
	})

	for {
		<-circuit.Done()
		slog.Info("closing cache")
		c.cache.Close()

		return
	}
}

func (c *Cache) Hits() uint64 {
	return c.cache.Metrics.Hits()
}

func (c *Cache) Misses() uint64 {
	return c.cache.Metrics.Misses()
}
