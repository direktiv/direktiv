package core

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Namespace represent a core domain object of direktiv.
type Namespace interface {
	GetID() uuid.UUID
	GetName() string
	CreatedAt() time.Time
	UpdatedAt() time.Time
}

// NamespaceStore responsible for fetching namespace from datastore.
type NamespaceStore interface {
	GetNamespace(ctx context.Context, id uuid.UUID) (Namespace, error)
	GetNamespaceByName(ctx context.Context, name string) (Namespace, error)
}
