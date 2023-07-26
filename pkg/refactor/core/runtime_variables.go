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

	NamespaceID  uuid.UUID
	WorkflowPath string
	InstanceID   uuid.UUID

	Name string

	Size     int
	MimeType string
	Data     []byte

	CreatedAt time.Time
	UpdatedAt time.Time
}

// RuntimeVariablesStore responsible for fetching and setting direktiv runtime variables from datastore.
//
//nolint:interfacebloat
type RuntimeVariablesStore interface {
	// GetByID gets a single runtime variable from store. if no record found,
	// it returns datastore.ErrNotFound error.
	GetByID(ctx context.Context, id uuid.UUID) (*RuntimeVariable, error)

	// GetByNamespaceAndName gets a single runtime variable from store by namespace ID and name. if no record found,
	// it returns datastore.ErrNotFound error.
	GetByNamespaceAndName(ctx context.Context, namespaceID uuid.UUID, name string) (*RuntimeVariable, error)

	// GetByWorkflowAndName gets a single runtime variable from store by namespace ID, workflow path, and name. if no record found,
	// it returns datastore.ErrNotFound error.
	GetByWorkflowAndName(ctx context.Context, namespaceID uuid.UUID, workflowPath, name string) (*RuntimeVariable, error)

	// GetByInstanceAndName gets a single runtime variable from store by instance ID and name. if no record found,
	// it returns datastore.ErrNotFound error.
	GetByInstanceAndName(ctx context.Context, instanceID uuid.UUID, name string) (*RuntimeVariable, error)

	// ListByInstanceID gets all runtime variable entries from store that are linked to specific instance id
	ListByInstanceID(ctx context.Context, instanceID uuid.UUID) ([]*RuntimeVariable, error)

	// ListByWorkflowPath gets all runtime variable entries from store that are linked to specific namespace & workflow path
	ListByWorkflowPath(ctx context.Context, namespaceID uuid.UUID, workflowPath string) ([]*RuntimeVariable, error)

	// ListByNamespaceID gets all runtime variable entries from store that are linked to specific namespace id
	ListByNamespaceID(ctx context.Context, namespaceID uuid.UUID) ([]*RuntimeVariable, error)

	// Set tries to update runtime variable data and mimetype fields or insert a new one if no matching variable to
	// update. Param variable should have one reference field set and name field set.
	Set(ctx context.Context, variable *RuntimeVariable) (*RuntimeVariable, error)

	// SetName updates a variable name. if no record found it returns datastore.ErrNotFound error.
	SetName(ctx context.Context, id uuid.UUID, name string) (*RuntimeVariable, error)

	// Delete removes the whole entry from store. if no record found it returns datastore.ErrNotFound error.
	Delete(ctx context.Context, id uuid.UUID) error

	// LoadData reads data field of a variable. if no record found it returns datastore.ErrNotFound error.
	LoadData(ctx context.Context, id uuid.UUID) ([]byte, error)
}

const RuntimeVariableNameRegexPattern = `^(([a-zA-Z][a-zA-Z0-9_\-\.]*[a-zA-Z0-9])|([a-zA-Z]))$`

var ErrInvalidRuntimeVariableName = errors.New("ErrInvalidRuntimeVariableName")
