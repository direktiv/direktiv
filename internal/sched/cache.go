package sched

import (
	"sort"
	"sync"
	"time"
)

type RuleCache struct {
	mu    sync.RWMutex
	items map[string]Rule // key: orderID
}

func NewRulesCache() *RuleCache {
	return &RuleCache{
		items: map[string]Rule{},
	}
}

func (c *RuleCache) Upsert(r *Rule) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// keep only the newest by Sequence
	if cur, ok := c.items[r.ID]; !ok || r.Sequence >= cur.Sequence {
		cp := r.Clone() // take ownership via clone
		c.items[r.ID] = *cp
	}
}

func (c *RuleCache) Snapshot(filterNamespace string) []*Rule {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]*Rule, 0, len(c.items))
	for _, v := range c.items {
		if v.Namespace != filterNamespace && filterNamespace != "" || v.DeletedAt != (time.Time{}) {
			continue
		}
		out = append(out, v.Clone())
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})

	return out
}
