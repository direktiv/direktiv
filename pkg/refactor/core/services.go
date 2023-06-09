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
	URL       string
	Data      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ServicesStore responsible for fetching and setting namespace services from datastore.
type ServicesStore interface {
	// GetByNamespaceAndName gets a single namespace service from the store. if no record found,
	// it returns datastore.ErrNotFound error.
	GetByNamespaceAndName(ctx context.Context, namespace uuid.UUID, name string) (*Service, error)

	// GetOneByURL gets a single namespace service from the store. if no record found,
	// it returns datastore.ErrNotFound error.
	GetOneByURL(ctx context.Context, url string) (*Service, error)

	// Create creates a service.
	Create(ctx context.Context, secret *Service) error

	// Update updates a service by id. if no record to update, it returns datastore.ErrNotFound error.
	Update(ctx context.Context, secret *Service) error

	// GetAll lists all services.
	GetAll(ctx context.Context) ([]*Service, error)

	// DeleteByName removes entries by name. if no record to delete, it returns datastore.ErrNotFound error.
	DeleteByName(ctx context.Context, name string) error
}
