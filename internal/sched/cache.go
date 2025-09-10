package sched

import (
	"sort"
	"sync"
)

type RulesCache struct {
	mu    sync.RWMutex
	items map[string]Rule // key: orderID
}

func NewRulesCache() *RulesCache {
	return &RulesCache{
		items: map[string]Rule{},
	}
}

func (c *RulesCache) Upsert(r *Rule) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// keep only the newest by Sequence
	if cur, ok := c.items[r.ID]; !ok || r.Sequence >= cur.Sequence {
		cp := r.Clone() // take ownership via clone
		c.items[r.ID] = *cp
	}
}

func (c *RulesCache) Snapshot(filterNamespace string) []*Rule {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]*Rule, 0, len(c.items))
	for _, v := range c.items {
		if v.Namespace != filterNamespace && filterNamespace != "" {
			continue
		}
		out = append(out, v.Clone())
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})

	return out
}
