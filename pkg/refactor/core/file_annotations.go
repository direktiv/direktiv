package core

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// FileAnnotations are extra data(string key -> string value) associated with files. We used it in direktiv to store
// user defined attributed to files.
type FileAnnotations struct {
	FileID uuid.UUID
	Data   FileAnnotationsData

	CreatedAt time.Time
	UpdatedAt time.Time
}

// FileAnnotationsData is the data part of file annotations.
type FileAnnotationsData map[string]string

// SetEntry sets key -> value into the data.
func (data FileAnnotationsData) SetEntry(key string, value string) FileAnnotationsData {
	if len(data) == 0 {
		return map[string]string{
			key: value,
		}
	}
	data[key] = value

	return data
}

// GetEntry returns empty string when key is not found.
func (data FileAnnotationsData) GetEntry(key string) string {
	if len(data) == 0 {
		return ""
	}
	val, ok := data[key]
	if !ok {
		return ""
	}

	return val
}

// RemoveEntry tries to remove key entry if found.
func (data FileAnnotationsData) RemoveEntry(key string) FileAnnotationsData {
	if len(data) == 0 {
		return data
	}
	delete(data, key)

	return data
}

var ErrFileAnnotationsNotSet = errors.New("ErrFileAnnotationsNotSet")

// FileAnnotationsStore responsible for fetching and setting file annotations from datastore.
type FileAnnotationsStore interface {
	// Get gets file annotation information from the store. if no record found,
	// it returns ErrFileAnnotationsNotSet error.
	Get(ctx context.Context, fileID uuid.UUID) (*FileAnnotations, error)

	// Set either creates (if not exists) file annotation information or updates the existing one.
	Set(ctx context.Context, annotations *FileAnnotations) error
}
