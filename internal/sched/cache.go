package sched

import (
	"sort"
	"sync"
)

type ConfigCache struct {
	mu    sync.RWMutex
	items map[string]Config // key: orderID
}

func NewConfigCache() *ConfigCache {
	return &ConfigCache{
		items: map[string]Config{},
	}
}

func (c *ConfigCache) Upsert(s Config) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// keep only the newest by Sequence (just in case messages race on the read side)
	if cur, ok := c.items[s.ID]; !ok || s.Sequence >= cur.Sequence {
		c.items[s.ID] = s
	}
}

func (c *ConfigCache) Snapshot(filterNamespace string) []Config {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]Config, 0, len(c.items))
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
