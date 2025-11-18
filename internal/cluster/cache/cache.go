package cache

import (
	"context"

	"github.com/direktiv/direktiv/internal/core"
)

type CacheNotify struct {
	Key    string
	Action CacheAction
}

type CacheAction string

const (
	CacheUpdate CacheAction = "update"
	CacheDelete CacheAction = "delete"
	CacheCreate CacheAction = "create"
	CacheClear  CacheAction = "clear"
)

type Manager interface {
	SecretsCache() Cache[core.Secret]
	FlowCache() Cache[core.TypescriptFlow]
	NamespaceCache() Cache[[]string]
}

type Cache[T any] interface {
	Get(key string, fetch func(...any) (T, error)) (T, error)
	Delete(key string)
	Set(key string, value T)
	Notify(ctx context.Context, notify CacheNotify)
}
