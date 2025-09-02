package datastore

import (
	"context"
	"errors"
	"time"
)

type Namespace struct {
	Name string `json:"name"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

var (
	ErrInvalidNamespaceName    = errors.New("ErrInvalidNamespaceName")
	ErrDuplicatedNamespaceName = errors.New("ErrDuplicatedNamespaceName")
)

// NamespacesStore responsible for fetching and setting namespaces from datastore.
type NamespacesStore interface {
	// GetByName gets a single namespace object from store. if no record found,
	// it returns datastore.ErrNotFound error.
	GetByName(ctx context.Context, name string) (*Namespace, error)

	// GetAll gets all namespaces from store.
	GetAll(ctx context.Context) ([]*Namespace, error)

	// Delete deletes a single namespace. if no record found,
	// it returns datastore.ErrNotFound error.
	Delete(ctx context.Context, name string) error

	// Create creates a new namespace. Returned errors could be ErrDuplicatedNamespaceName when namespace name is
	// already exists or ErrInvalidNamespaceName or when namespace name is invalid, too short or too long.
	Create(ctx context.Context, namespace *Namespace) (*Namespace, error)
}
