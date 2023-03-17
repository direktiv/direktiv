package filestore

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
)

type RevisionTags []string

func (t RevisionTags) String() string {
	if len(t) == 0 {
		return ""
	}

	return strings.Join(t, ",")
}

func ParseRevisionTags(tags string) RevisionTags {
	return strings.Split(tags, ",")
}

type Revision struct {
	ID        uuid.UUID
	Tags      string
	IsCurrent bool
	Data      []byte

	FileID uuid.UUID

	CreatedAt time.Time
	UpdatedAt time.Time
}

type RevisionQuery interface {
	GetData(ctx context.Context) (io.ReadCloser, error)
	SetCurrent(ctx context.Context) (*Revision, error)
	SetTags(ctx context.Context, tags RevisionTags) (*Revision, error)
	Delete(ctx context.Context, force bool) error
}
