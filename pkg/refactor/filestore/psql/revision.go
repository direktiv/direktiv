package psql

import (
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/google/uuid"
)

type Revision struct {
	ID        uuid.UUID
	Tags      string
	IsCurrent bool
	Data      []byte

	FileID uuid.UUID

	CreatedAt time.Time
	UpdatedAt time.Time
}

var _ filestore.Revision = &Revision{} // Ensures Revision struct conforms to filestore.Revision interface.

func (r *Revision) GetTags() filestore.RevisionTags {
	return filestore.ParseRevisionTags(r.Tags)
}

func (r *Revision) GetIsCurrent() bool {
	return false
}

func (r *Revision) GetCreatedAt() time.Time {
	return r.CreatedAt
}

func (r *Revision) GetUpdatedAt() time.Time {
	return r.UpdatedAt
}

func (r *Revision) GetID() uuid.UUID {
	return r.ID
}

func (r *Revision) GetFileID() uuid.UUID {
	return r.FileID
}
