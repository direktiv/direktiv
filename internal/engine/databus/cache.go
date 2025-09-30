package databus

import (
	"sort"
	"sync"

	"github.com/direktiv/direktiv/internal/engine"
	"github.com/google/uuid"
)

type StatusCache struct {
	mu    sync.RWMutex
	items map[uuid.UUID]engine.InstanceStatus // key: orderID
}

func NewStatusCache() *StatusCache {
	return &StatusCache{
		items: map[uuid.UUID]engine.InstanceStatus{},
	}
}

func (c *StatusCache) Upsert(s *engine.InstanceStatus) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// keep only the newest by HistorySequence
	if cur, ok := c.items[s.InstanceID]; !ok || s.HistorySequence >= cur.HistorySequence {
		cp := s.Clone()
		c.items[s.InstanceID] = *cp
	}
}

func (c *StatusCache) Snapshot(filterNamespace string, filterInstanceID uuid.UUID) []*engine.InstanceStatus {
	return c.SnapshotPage(filterNamespace, filterInstanceID, 0, 0)
}

func (c *StatusCache) SnapshotPage(filterNamespace string, filterInstanceID uuid.UUID, limit int, offset int) []*engine.InstanceStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]*engine.InstanceStatus, 0, len(c.items))

	for _, v := range c.items {
		if v.InstanceID != filterInstanceID && filterInstanceID != uuid.Nil {
			continue
		}
		if v.Namespace != filterNamespace && filterNamespace != "" {
			continue
		}
		if offset > 0 {
			offset--
			continue
		}
		out = append(out, v.Clone())
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})

	return out
}

func (c *StatusCache) DeleteNamespace(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range c.items {
		if name == v.Namespace {
			delete(c.items, k)
		}
	}
}
