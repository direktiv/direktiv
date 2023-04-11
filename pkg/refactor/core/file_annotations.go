package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type FileAnnotations struct {
	FileID uuid.UUID
	Data   FileAnnotationsData

	CreatedAt time.Time
	UpdatedAt time.Time
}

type FileAnnotationsData []byte

func NewFileAnnotationsData(list map[string]string) FileAnnotationsData {
	if len(list) == 0 {
		return []byte("{}")
	}
	jsonStr, err := json.Marshal(list)
	if err != nil {
		panic(fmt.Sprintf("logic error, marshalling FileAnnotationsData with value: %v, got error: %s", list, err))
	}

	return jsonStr
}

func (data FileAnnotationsData) SetEntry(key string, value string) FileAnnotationsData {
	if len(data) == 0 {
		data = []byte("{}")
	}
	list := map[string]string{}
	err := json.Unmarshal(data, &list)
	if err != nil {
		panic(fmt.Sprintf("logic error, unmarshalling FileAnnotationsData with val: %s, got error: %s", data, err))
	}
	list[key] = value

	return NewFileAnnotationsData(list)
}

func (data FileAnnotationsData) GetEntry(key string) string {
	if len(data) == 0 {
		data = []byte("{}")
	}
	list := map[string]string{}
	err := json.Unmarshal(data, &list)
	if err != nil {
		panic(fmt.Sprintf("logic error, unmarshalling FileAnnotationsData with val: %s, got error: %s", data, err))
	}
	val, ok := list[key]
	if !ok {
		return ""
	}

	return val
}

func (data FileAnnotationsData) RemoveEntry(key string) FileAnnotationsData {
	if len(data) == 0 {
		data = []byte("{}")
	}
	list := map[string]string{}
	err := json.Unmarshal(data, &list)
	if err != nil {
		panic(fmt.Sprintf("logic error, unmarshalling FileAnnotationsData with val: %s, got error: %s", data, err))
	}
	delete(list, key)

	return NewFileAnnotationsData(list)
}

var ErrFileAnnotationsNotSet = errors.New("ErrFileAnnotationsNotSet")

// FileAnnotationsStore responsible for fetching file annotations from datastore.
type FileAnnotationsStore interface {
	Get(ctx context.Context, fileID uuid.UUID) (*FileAnnotations, error)
	Set(ctx context.Context, annotations *FileAnnotations) error
}
