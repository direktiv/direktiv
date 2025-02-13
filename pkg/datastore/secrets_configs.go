package datastore

import (
	"context"
	"time"
)

type SecretsConfigs struct {
	Namespace string

	Configuration []byte

	CreatedAt time.Time
	UpdatedAt time.Time
}

type SecretsConfigsStore interface {
	// Set upserts a SecretsConfigs in the store.
	Set(ctx context.Context, config *SecretsConfigs) error

	// Get gets SecretsConfigs by namespace from the store.
	Get(ctx context.Context, namespace string) (*SecretsConfigs, error)

	// NOTE: we want there to always be a config for a namespace, so we rely on cascade for deletion.
}
