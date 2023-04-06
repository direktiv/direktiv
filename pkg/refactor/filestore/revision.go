package filestore

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
)

type RevisionTags string

func (tags RevisionTags) AddTag(tag string) RevisionTags {
	tag = strings.TrimSpace(tag)
	if strings.Contains(string(tags), tag) {
		return tags
	}

	return RevisionTags(string(tags) + "," + tag)
}

func (tags RevisionTags) RemoveTag(tag string) RevisionTags {
	tag = strings.TrimSpace(tag)

	newTags := strings.Replace(string(tags), ","+tag, "", 1)
	newTags = strings.Replace(newTags, tag+",", "", 1)

	return RevisionTags(newTags)
}

func (tags RevisionTags) List() []string {
	return strings.Split(string(tags), ",")
}

type Revision struct {
	ID        uuid.UUID
	Tags      RevisionTags
	IsCurrent bool
	Data      []byte
	Checksum  string

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
