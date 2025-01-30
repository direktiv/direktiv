package metastore

import "context"

type MetricsStore interface {
	Init(ctx context.Context) error
	Get(ctx context.Context, label string, options MetricsQueryOptions) ([]map[string]any, error)
	GetAll(ctx context.Context, limit int) ([]string, error)
}

type MetricsQueryOptions struct {
	Limit int // Maximum number of entries to return
}
