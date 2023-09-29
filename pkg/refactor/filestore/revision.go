package filestore

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

// RevisionTags is a comma separated string of tags that refers to the revision.
type RevisionTags string

// AddTag adds a new tag to the tags string.
func (tags RevisionTags) AddTag(tag string) RevisionTags {
	tag = strings.Trim(tag, ",")
	tag = strings.Split(tag, ",")[0]
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return tags
	}

	tagsStr := string(tags)
	tagsStr = strings.TrimSpace(tagsStr)
	tagsStr = strings.Trim(tagsStr, ",")

	if strings.Contains(","+tagsStr+",", ","+tag+",") {
		return tags
	}

	newTags := tagsStr + "," + tag
	newTags = strings.Trim(newTags, ",")

	return RevisionTags(newTags)
}

// RemoveTag removes tag from the tags string.
func (tags RevisionTags) RemoveTag(tag string) RevisionTags {
	tag = strings.Trim(tag, ",")
	tag = strings.Split(tag, ",")[0]
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return tags
	}

	tagsStr := string(tags)
	tagsStr = strings.TrimSpace(tagsStr)
	tagsStr = strings.Trim(tagsStr, ",")

	newTags := strings.Replace(","+tagsStr+",", ","+tag+",", ",", 1)
	newTags = strings.Trim(newTags, ",")

	return RevisionTags(newTags)
}

func (tags RevisionTags) List() []string {
	if string(tags) == "" {
		return []string{}
	}

	return strings.Split(string(tags), ",")
}

// Revision is a snapshot of a file in the filestore, every file has at least one revision which is the current
// revision. File revisions is not applicable to directory file type.
type Revision struct {
	ID uuid.UUID

	// Tags is a comma separated string of tags that refer for a revision.
	Tags RevisionTags

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

	// SetTags set tags of a revision.
	SetTags(ctx context.Context, tags RevisionTags) error

	// Delete deletes file revision.
	Delete(ctx context.Context) error
}
