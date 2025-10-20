package databus

import (
	"sync"

	"github.com/direktiv/direktiv/internal/api/filter"
	"github.com/direktiv/direktiv/internal/engine"
	"github.com/google/uuid"
)

type StatusCache struct {
	mu    sync.RWMutex
	items []engine.InstanceStatus
	index map[uuid.UUID]int
}

func NewStatusCache() *StatusCache {
	return &StatusCache{
		items: make([]engine.InstanceStatus, 0),
		index: make(map[uuid.UUID]int),
	}
}

func (c *StatusCache) Upsert(s *engine.InstanceStatus) {
	c.mu.Lock()
	defer c.mu.Unlock()
	i, ok := c.index[s.InstanceID]

	// if not found, add it
	if !ok || i < 0 || i >= len(c.items) {
		cp := s.Clone()
		c.items = append(c.items, *cp)
		c.index[s.InstanceID] = len(c.items) - 1

		return
	}

	// here we need to update only if HistorySequence is newer
	v := &c.items[i]
	if v.HistorySequence < s.HistorySequence {
		cp := s.Clone()
		c.items[i] = *cp
	}
}

func (c *StatusCache) Snapshot() []*engine.InstanceStatus {
	res, _ := c.SnapshotPage(0, 0, nil)
	return res
}

func (c *StatusCache) SnapshotPage(limit int, offset int, filters filter.Values) ([]*engine.InstanceStatus, int) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	out := make([]*engine.InstanceStatus, 0, len(c.items))

	total := 0
	for i := len(c.items) - 1; i >= 0; i-- {
		v := &c.items[i]
		if !filters.Match("instanceID", v.InstanceID.String()) {
			continue
		}
		if !filters.Match("namespace", v.Namespace) {
			continue
		}

		total++
		if offset > 0 {
			offset--
			continue
		}
		if limit > 0 && len(out) >= limit {
			continue
		}
		out = append(out, v.Clone())
	}

	return out, total
}

func (c *StatusCache) DeleteNamespace(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	cp := make([]engine.InstanceStatus, 0, len(c.items))
	index := make(map[uuid.UUID]int)
	for _, v := range c.items {
		if name != v.Namespace {
			cp = append(cp, v)
			index[v.InstanceID] = len(cp) - 1
		}
	}
	c.items = cp
	c.index = index
}
