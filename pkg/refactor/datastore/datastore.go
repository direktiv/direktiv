package datastore

import (
	"context"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/mirror"
)

type BaseStore interface {
	Mirror() mirror.Store
	FileAttributes() core.FileAttributesStore
}

type Store interface {
	BaseStore
	Begin(ctx context.Context) (StoreTx, error)
}

type StoreTx interface {
	BaseStore
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
