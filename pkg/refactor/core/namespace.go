package core

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Namespace represent a core domain object of direktiv.
type Namespace struct {
	ID   uuid.UUID
	Name string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// NamespaceStore responsible for fetching namespace from datastore.
type NamespaceStore interface {
	Create(ctx context.Context, namespace *Namespace) error
	Get(ctx context.Context, id uuid.UUID) (*Namespace, error)
	GetByName(ctx context.Context, name string) (*Namespace, error)
	Delete(ctx context.Context, name string) error
}
