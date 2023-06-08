package core

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	ID          uuid.UUID
	NamespaceID uuid.UUID

	Name      string
	Url       string
	Data      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ServicesStore responsible for fetching and setting namespace services from datastore.
type ServicesStore interface {
	// GetByName gets a single namespace service from the store. if no record found,
	// it returns datastore.ErrNotFound error.
	GetByName(ctx context.Context, namespace uuid.UUID, name string) (*Service, error)

	// GetByUrl gets a single namespace service from the store. if no record found,
	// it returns datastore.ErrNotFound error.
	GetByUrl(ctx context.Context, name string) (*Service, error)

	// Create creates a service.
	Create(ctx context.Context, secret *Service) error

	// Update creates a service.
	Update(ctx context.Context, secret *Service) error

	// GetAll lists all namespace secrets.
	GetAll(ctx context.Context) ([]*Service, error)

	// Delete removes a specific secret by name.
	Delete(ctx context.Context, name string) error
}
