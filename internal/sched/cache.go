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

func (c *RulesCache) Upsert(s Rule) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// keep only the newest by Sequence
	if cur, ok := c.items[s.ID]; !ok || s.Sequence >= cur.Sequence {
		c.items[s.ID] = s
	}
}

func (c *RulesCache) Snapshot(filterNamespace string) []Rule {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]Rule, 0, len(c.items))
	for _, v := range c.items {
		if v.Namespace != filterNamespace && filterNamespace != "" {
			continue
		}
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})

	return out
}
