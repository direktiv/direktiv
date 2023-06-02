package core

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// RuntimeVariable are direktiv runtime variables that hold data, workflows performs getting and setting on these
// data, RuntimeVariables also preserve state across multiple workflow runs.
type RuntimeVariable struct {
	ID uuid.UUID

	NamespaceID uuid.UUID
	WorkflowID  uuid.UUID
	InstanceID  uuid.UUID

	Name string

	Size     int
	MimeType string
	Data     []byte

	CreatedAt time.Time
	UpdatedAt time.Time
}

// RuntimeVariablesStore responsible for fetching and setting direktiv runtime variables from datastore.
type RuntimeVariablesStore interface {
	// GetByID gets a single runtime variable from store. if no record found,
	// it returns datastore.ErrNotFound error.
	GetByID(ctx context.Context, id uuid.UUID) (*RuntimeVariable, error)

	// GetByReferenceAndName gets a single runtime variable from store by reference id and name. if no record found,
	// it returns datastore.ErrNotFound error.
	// Each runtime variable is linked to a namespace, workflow, or instance. Param referenceID specifying the id of
	// the referencing object.
	GetByReferenceAndName(ctx context.Context, referenceID uuid.UUID, name string) (*RuntimeVariable, error)

	// ListByInstanceID gets all runtime variable entries from store that are linked to specific instance id
	// if no record found, it returns datastore.ErrNotFound error.
	ListByInstanceID(ctx context.Context, instanceID uuid.UUID) ([]*RuntimeVariable, error)

	// ListByWorkflowID gets all runtime variable entries from store that are linked to specific workflow id
	// if no record found, it returns datastore.ErrNotFound error.
	ListByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]*RuntimeVariable, error)

	// ListByNamespaceID gets all runtime variable entries from store that are linked to specific namespace id
	// if no record found, it returns datastore.ErrNotFound error.
	ListByNamespaceID(ctx context.Context, namespaceID uuid.UUID) ([]*RuntimeVariable, error)

	// Set tries to update runtime variable data and mimetype fields or insert a new one if no matching variable to
	// update. Param variable should have one reference field set and name field set.
	Set(ctx context.Context, variable *RuntimeVariable) (*RuntimeVariable, error)

	// SetName updates a variable name.
	SetName(ctx context.Context, id uuid.UUID, name string) (*RuntimeVariable, error)

	// Delete removes the whole entry from store.
	Delete(ctx context.Context, id uuid.UUID) error

	// LoadData reads data field of a variable.
	LoadData(ctx context.Context, id uuid.UUID) ([]byte, error)
}

const RuntimeVariableNameRegexPattern = `^(([a-zA-Z][a-zA-Z0-9_\-\.]*[a-zA-Z0-9])|([a-zA-Z]))$`

var ErrInvalidRuntimeVariableName = errors.New("ErrInvalidRuntimeVariableName")
