package instancestore

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound = errors.New("not found")
)

type InstanceData struct {
	ID              uuid.UUID
	NamespaceID     uuid.UUID
	WorkflowID      uuid.UUID
	RevisionID      uuid.UUID
	RootInstanceID  uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
	EndedAt         *time.Time
	Deadline        *time.Time
	Status          string // TODO: alan, turn this into an enum so we can do arithmetic comparisons on it
	CalledAs        string
	ErrorCode       string
	Invoker         string
	Definition      []byte // TODO: alan, we should strip comments
	Settings        []byte
	DescentInfo     []byte
	TelemetryInfo   []byte
	RuntimeInfo     []byte
	ChildrenInfo    []byte
	Input           []byte
	LiveData        []byte
	TemporaryMemory []byte
	Output          []byte
	ErrorMessage    []byte
	Metadata        []byte
	// TODO: alan, should we consider compressing these binary fields?
}

type CreateInstanceDataArgs struct {
	ID             uuid.UUID
	NamespaceID    uuid.UUID
	WorkflowID     uuid.UUID
	RevisionID     uuid.UUID
	RootInstanceID uuid.UUID
	Invoker        string
	CalledAs       string
	Definition     []byte
	Input          []byte
	TelemetryInfo  []byte
	Settings       []byte
	DescentInfo    []byte
}

// TODO: alan
type UpdateInstanceDataArgs struct {
}

type InstanceDataQuery interface {
	// TODO: alan, implement UpdateInstanceData
	UpdateInstanceData(ctx context.Context, args *UpdateInstanceDataArgs) (*InstanceData, error)

	// TODO: alan, implement more fine-grained setters?
	// - UpdateInstanceTransition
	// - UpdateInstanceYield
	// - UpdateInstanceTerminateTerminate

	GetEverything(ctx context.Context) (*InstanceData, error)
	GetSummary(ctx context.Context) (*InstanceData, error)

	// TODO: alan, implement more fine-grained getters?
	// - get everything except blobs that could be large AND are almost never read by the engine, only written: (input, output, error_message, metadata)
	// - what about things that are only sometimes relevent to the engine? (children_info)
}

type Store interface {
	// TODO: alan, discuss with Yassir if the namespace ID should be an argument here
	ForInstanceID(id uuid.UUID) InstanceDataQuery
	CreateInstanceData(ctx context.Context, args *CreateInstanceDataArgs) (*InstanceData, error)
	// TODO: alan, pagination/filtering on this list
	GetNamespaceInstances(ctx context.Context, nsID uuid.UUID) ([]*InstanceData, error)

	// TODO: alan, GetHangingInstances :: query to get a list of all instances where deadline has been exceeded and status is unfinished
	// TODO: alan, DeleteOldInstances :: query to delete all instances that have terminated and end_at a long time ago
	// TODO: alan, CheckForParallelCron :: query to try and detect if another machine in a HA environment may have already triggered an instance that we're just about to create ourselves
}
