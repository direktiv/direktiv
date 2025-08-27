package cache

import (
	"encoding/json"
	"log/slog"
	"time"
	"unsafe"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/direktiv/direktiv/pkg/core"
)

type cacheMessage struct {
	Hostname string
	Key      string
}

type Cache struct {
	cache    *ristretto.Cache[string, any]
	bus      core.PubSub
	hostname string
}

func NewCache(bus core.PubSub, hostname string, enableMetrics bool) (core.Cache, error) {
	cache, err := ristretto.NewCache(&ristretto.Config[string, any]{
		NumCounters: 10000000,
		MaxCost:     1073741824,
		BufferItems: 64,
		Metrics:     enableMetrics,
	})
	if err != nil {
		return nil, err
	}

	return &Cache{
		bus:      bus,
		cache:    cache,
		hostname: hostname,
	}, nil
}

func (c *Cache) SetTTL(key string, value any, ttl int) {
	if ttl == 0 {
		c.cache.Set(key, value, int64(unsafe.Sizeof(value)))
	} else {
		c.cache.SetWithTTL(key, value, int64(unsafe.Sizeof(value)),
			time.Duration(ttl)*time.Second)
	}

	c.cache.Wait()
	c.publish(key)
}

func (c *Cache) Set(key string, value any) {
	c.SetTTL(key, value, 0)
}

func (c *Cache) Delete(key string) {
	c.cache.Del(key)
	c.publish(key)
}

func (c *Cache) Get(key string) (any, bool) {
	return c.cache.Get(key)
}

func (c *Cache) Run(circuit *core.Circuit) {
	c.subscribe()
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

func (c *Cache) publish(key string) {
	cm := &cacheMessage{
		Hostname: c.hostname,
		Key:      key,
	}
	b, err := json.Marshal(cm)
	if err != nil {
		slog.Error("can not publish cache", slog.Any("error", err))
	}

	err = c.bus.Publish(core.CacheDeleteEvent, b)
	if err != nil {
		slog.Error("can not publish cache", slog.Any("error", err))
	}
}

func (c *Cache) subscribe() {
	c.bus.Subscribe(core.CacheDeleteEvent, func(data []byte) {
		var cm cacheMessage
		err := json.Unmarshal(data, &cm)
		if err != nil {
			return
		}

		if cm.Hostname != c.hostname {
			c.cache.Del(cm.Key)
		}
	})
}
