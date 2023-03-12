package psql

import (
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/google/uuid"
)

type Root struct {
	ID uuid.UUID

	Files []File `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

var _ filestore.Root = &Root{} // Ensures Root struct conforms to filestore.Root interface.

func (r *Root) GetCreatedAt() time.Time {
	return r.CreatedAt
}

func (r *Root) GetUpdatedAt() time.Time {
	return r.UpdatedAt
}

func (r *Root) GetID() uuid.UUID {
	return r.ID
}
