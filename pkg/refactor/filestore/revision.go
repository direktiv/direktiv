package filestore

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Revision is a snapshot of a file in the filestore, every file has at least one revision which is the current
// revision. File revisions is not applicable to directory file type.
type Revision struct {
	ID uuid.UUID

	// IsCurrent flags if a revision is a current file revision.
	IsCurrent bool
	Data      []byte
	Checksum  string

	FileID uuid.UUID

	CreatedAt time.Time
	UpdatedAt time.Time
}

// RevisionQuery performs different queries associated to a file revision.
type RevisionQuery interface {
	// GetData gets data of a revision.
	GetData(ctx context.Context) ([]byte, error)

	// SetCurrent sets a revision to be the current one.
	SetCurrent(ctx context.Context) (*Revision, error)

	// Delete deletes file revision.
	Delete(ctx context.Context) error
}
