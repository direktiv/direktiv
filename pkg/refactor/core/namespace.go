package core

import (
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
	GetNamespace(id uuid.UUID) (Namespace, error)
	GetNamespaceByName(name string) (Namespace, error)
}
