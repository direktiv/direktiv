package datastore

import (
	"context"
	"errors"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/google/uuid"
)

type Namespace struct {
	ID uuid.UUID `json:"-"`

	Name string `json:"name"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (ns *Namespace) WithTags(ctx context.Context) context.Context {
	tags, ok := ctx.Value(core.TagsKey).([]interface{})
	if !ok {
		tags = make([]interface{}, 0)
	}
	tags = append(tags, "namespace", ns.Name)

	return context.WithValue(ctx, core.TagsKey, tags)
}

func (ns *Namespace) GetAttributes() map[string]string {
	return map[string]string{
		"namespace":    ns.Name,
		"namespace-id": ns.ID.String(),
	}
}

var (
	ErrInvalidNamespaceName    = errors.New("ErrInvalidNamespaceName")
	ErrDuplicatedNamespaceName = errors.New("ErrDuplicatedNamespaceName")
)

// NamespacesStore responsible for fetching and setting namespaces from datastore.
type NamespacesStore interface {
	// GetByID gets a single namespace object from store. if no record found,
	// it returns datastore.ErrNotFound error.
	GetByID(ctx context.Context, id uuid.UUID) (*Namespace, error)

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
