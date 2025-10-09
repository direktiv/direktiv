package databus

import (
	"sync"

	"github.com/direktiv/direktiv/internal/engine"
	"github.com/google/uuid"
)

type HistoryCache struct {
	mu    sync.RWMutex
	items []engine.InstanceEvent
}

func NewHistoryCache() *HistoryCache {
	return &HistoryCache{
		items: make([]engine.InstanceEvent, 0),
	}
}

func (c *HistoryCache) Insert(s *engine.InstanceEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// if the sequence is less than the last one, we don't need to add it as it could be a duplicate
	if len(c.items) > 0 && s.Sequence <= c.items[len(c.items)-1].Sequence {
		return
	}
	cp := s.Clone()
	c.items = append(c.items, *cp)
}

func (c *HistoryCache) Snapshot(namespace string, instanceID uuid.UUID) []*engine.InstanceEvent {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var out []*engine.InstanceEvent
	for i := range c.items {
		v := &c.items[i]
		if v.InstanceID == instanceID && v.Namespace == namespace {
			out = append(out, v.Clone())
		}
	}

	return out
}

func (c *HistoryCache) DeleteNamespace(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	newItems := make([]engine.InstanceEvent, 0, len(c.items))
	for i := range c.items {
		v := c.items[i]
		if v.Namespace != name {
			newItems = append(newItems, v)
		}
	}
	c.items = newItems
}
