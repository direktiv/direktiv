package core

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

type FileAttributesValue string

func newFileAttributesValueWithExcludes(list []string, excludes []string) FileAttributesValue {
	// create a map with all the values as key
	uniqMap := make(map[string]bool)
	for _, v := range list {
		v = strings.TrimSpace(v)
		if len(v) > 0 {
			uniqMap[v] = true
		}
	}

	// remove excludes from the map
	for _, v := range excludes {
		v = strings.TrimSpace(v)
		delete(uniqMap, v)
	}

	// turn the map keys into a slice
	uniqSlice := make([]string, 0, len(uniqMap))
	for v := range uniqMap {
		uniqSlice = append(uniqSlice, v)
	}
	sort.Strings(uniqSlice)

	return FileAttributesValue(strings.Join(uniqSlice, ","))
}

func NewFileAttributesValue(list []string) FileAttributesValue {
	return newFileAttributesValueWithExcludes(list, []string{})
}

type FileAttributes struct {
	FileID uuid.UUID
	Value  FileAttributesValue

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (a *FileAttributes) Add(attributes []string) *FileAttributes {
	oldList := strings.Split(string(a.Value), ",")
	newValue := NewFileAttributesValue(append(oldList, attributes...))

	return &FileAttributes{
		FileID:    a.FileID,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
		Value:     newValue,
	}
}

func (a *FileAttributes) Remove(attributes []string) *FileAttributes {
	oldList := strings.Split(string(a.Value), ",")

	newValue := newFileAttributesValueWithExcludes(oldList, attributes)

	return &FileAttributes{
		FileID:    a.FileID,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
		Value:     newValue,
	}
}

var ErrFileAttributesNotSet = errors.New("ErrFileAttributesNotSet")

// FileAttributesStore responsible for fetching file attributes from datastore.
type FileAttributesStore interface {
	Get(ctx context.Context, fileID uuid.UUID) (*FileAttributes, error)
	Set(ctx context.Context, fileAttributes *FileAttributes) error
}
