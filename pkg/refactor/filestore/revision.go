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
	tagsStr := string(tags)
	tagsStr = strings.TrimSpace(tagsStr)
	tagsStr = strings.Trim(tagsStr, ",")

	tag = strings.TrimSpace(tag)
	tag = strings.Trim(tag, ",")

	if strings.Contains(tagsStr, tag) {
		return RevisionTags(tagsStr)
	}

	newTags := tagsStr + "," + tag
	newTags = strings.Trim(newTags, ",")

	return RevisionTags(newTags)
}

func (tags RevisionTags) RemoveTag(tag string) RevisionTags {
	tagsStr := string(tags)
	tagsStr = strings.TrimSpace(tagsStr)
	tagsStr = strings.Trim(tagsStr, ",")

	tag = strings.TrimSpace(tag)
	tag = strings.Trim(tag, ",")

	newTags := strings.Replace(tagsStr, ","+tag, "", 1)
	newTags = strings.Replace(newTags, tag+",", "", 1)
	newTags = strings.Replace(newTags, tag, "", 1)

	return RevisionTags(newTags)
}

func (tags RevisionTags) List() []string {
	if string(tags) == "" {
		return []string{}
	}

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
