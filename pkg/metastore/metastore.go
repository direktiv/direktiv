package metastore

import (
	"context"
)

type Store interface {
	EventsStore() EventsStore
	LogStore() LogStore
	MetricsStore() MetricsStore
	TimelineStore() TimelineStore
	GetMapping(ctx context.Context, index string) (map[string]interface{}, error)
}
