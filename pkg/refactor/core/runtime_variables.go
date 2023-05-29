package core

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// RuntimeVariable are direktiv runtime variables that are hold data, workflows performs getting and setting on these
// data, RuntimeVariables also preserve state across multiple workflow runs.
type RuntimeVariable struct {
	ID uuid.UUID

	NamespaceID uuid.UUID
	WorkflowID  uuid.UUID
	InstanceID  uuid.UUID

	Scope string

	Name string

	Size     int
	Hash     string
	MimeType string
	Data     []byte

	CreatedAt time.Time
	UpdatedAt time.Time
}

type RuntimeVariablesList []*RuntimeVariable

// RuntimeVariablesStore responsible for fetching and setting direktiv runtime variables from datastore.
type RuntimeVariablesStore interface {
	// GetByID gets a single runtime variable from store. if no record found,
	// it returns core.ErrNotFound error.
	GetByID(ctx context.Context, id uuid.UUID) (*RuntimeVariable, error)

	// ListByInstanceID gets all runtime variable entries from store that are linked to specific instance id
	// if no record found, it returns core.ErrNotFound error.
	ListByInstanceID(ctx context.Context, instanceID uuid.UUID) (RuntimeVariablesList, error)

	// ListByWorkflowID gets all runtime variable entries from store that are linked to specific workflow id
	// if no record found, it returns core.ErrNotFound error.
	ListByWorkflowID(ctx context.Context, workflowID uuid.UUID) (RuntimeVariablesList, error)

	// ListByNamespaceID gets all runtime variable entries from store that are linked to specific namespace id
	// if no record found, it returns store.ErrNotFound error.
	ListByNamespaceID(ctx context.Context, namespaceID uuid.UUID) (RuntimeVariablesList, error)

	Set(ctx context.Context, variable *RuntimeVariable) (*RuntimeVariable, error)

	SetName(ctx context.Context, id uuid.UUID, name string) (*RuntimeVariable, error)

	// Delete removes the whole entry from store.
	Delete(ctx context.Context, id uuid.UUID) error

	LoadData(ctx context.Context, id uuid.UUID) ([]byte, error)
}

func (l RuntimeVariablesList) FilterByName(name string) *RuntimeVariable {
	for i := range l {
		if l[i].Name == name {
			return l[i]
		}
	}

	return nil
}
