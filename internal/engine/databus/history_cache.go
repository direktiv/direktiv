package databus

import (
	"sync"

	"github.com/direktiv/direktiv/internal/engine"
	"github.com/google/uuid"
)

type HistoryCache struct {
	mu    sync.RWMutex
	items []engine.InstanceEvent // key: orderID
}

func NewHistoryCache() *HistoryCache {
	return &HistoryCache{
		items: make([]engine.InstanceEvent, 0),
	}
}

// TODO: fix immutablility
func (c *HistoryCache) Insert(s *engine.InstanceEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = append(c.items, *s)

}

func (c *HistoryCache) Snapshot(namespace string, instanceID uuid.UUID) []*engine.InstanceEvent {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var out []*engine.InstanceEvent
	for i := range c.items {
		v := &c.items[i]
		if v.InstanceID == instanceID && v.Namespace == namespace {
			out = append(out, v)
		}
	}

	return out
}

func (c *HistoryCache) DeleteNamespace(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	cp := make([]engine.InstanceEvent, 0, len(c.items))
	for i := range c.items {
		v := &c.items[i]
		if v.Namespace != name {
			cp[i] = *v
		}
	}
	c.items = cp
}
