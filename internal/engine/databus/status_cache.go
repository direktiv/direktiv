package databus

import (
	"sync"
	"time"

	"github.com/direktiv/direktiv/internal/api/filter"
	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/internal/engine"
	"github.com/google/uuid"
)

type StatusCache struct {
	mu    sync.RWMutex
	items []engine.InstanceEvent
	index map[uuid.UUID]int
}

func NewStatusCache() *StatusCache {
	return &StatusCache{
		items: make([]engine.InstanceEvent, 0),
		index: make(map[uuid.UUID]int),
	}
}

func (c *StatusCache) Insert(s *engine.InstanceEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// if the sequence is less than the last one, we don't need to add it as it could be a duplicate
	if len(c.items) > 0 && s.Sequence <= c.items[len(c.items)-1].Sequence {
		return
	}
	cp := s.Clone()
	c.items = append(c.items, *cp)
	c.index[s.InstanceID] = len(c.items) - 1
}

func (c *StatusCache) Upsert(s *engine.InstanceEvent) {
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
	if v.Sequence < s.Sequence {
		cp := s.Clone()
		c.items[i] = *cp
	}
}

func (c *StatusCache) Snapshot(filters filter.Values) []*engine.InstanceEvent {
	res, _ := c.SnapshotPage(0, 0, filters)
	return res
}

func (c *StatusCache) SnapshotPage(limit int, offset int, filters filter.Values) ([]*engine.InstanceEvent, int) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	out := make([]*engine.InstanceEvent, 0, len(c.items))

	total := 0
	for i := len(c.items) - 1; i >= 0; i-- {
		v := &c.items[i]
		if !filters.Match("instanceID", v.InstanceID.String()) {
			continue
		}
		if !filters.Match("namespace", v.Namespace) {
			continue
		}
		if !filters.Match("status", string(v.State)) {
			continue
		}
		if !filters.Match("createdAt", v.CreatedAt.Format(time.RFC3339Nano)) {
			continue
		}
		workflowPath := v.Metadata[core.EngineMappingPath]
		if !filters.Match("metadata_"+core.EngineMappingPath, workflowPath) {
			continue
		}
		invokerType := v.Metadata[engine.LabelInvokerType]
		if !filters.Match("metadata_"+engine.LabelInvokerType, invokerType) {
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
	cp := make([]engine.InstanceEvent, 0, len(c.items))
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
