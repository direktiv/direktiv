package instancestore

import (
	"context"
	sqldriver "database/sql/driver"
	"errors"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
)

// Exported errors.
var (
	ErrNotFound     = errors.New("not found")
	ErrParallelCron = errors.New("a parallel cron already exists")
	ErrBadListOpts  = errors.New("unsupported list option")
	ErrNoMessages   = errors.New("no messages")
)

// InstanceStatus enum allows us to perform arithmetic comparisons on the database.
type InstanceStatus int

const (
	InstanceStatusPending InstanceStatus = iota + 1
	InstanceStatusComplete
	InstanceStatusFailed
	InstanceStatusCrashed
	InstanceStatusCancelled
)

var instanceStatusStrings = []string{
	util.InstanceStatusPending,
	util.InstanceStatusComplete,
	util.InstanceStatusFailed,
	util.InstanceStatusCrashed,
	util.InstanceStatusCancelled,
}

func (status InstanceStatus) String() string {
	return instanceStatusStrings[status-1]
}

func (status *InstanceStatus) Scan(src any) error {
	k, ok := src.(int64)
	if !ok {
		return errors.New("unknown instance status type")
	}

	*status = InstanceStatus(k)

	return nil
}

func (status InstanceStatus) Valuer() (sqldriver.Value, error) {
	return int64(status), nil
}

func InstanceStatusFromString(s string) (InstanceStatus, error) {
	for idx, x := range instanceStatusStrings {
		if s == x {
			return InstanceStatus(idx + 1), nil
		}
	}

	return InstanceStatus(0), fmt.Errorf("invalid instance status '%s' (expect one of %v)", s, instanceStatusStrings)
}

// InvokerCron defined here so drivers can use it in their AssertNoParallelCron implementation.
const InvokerCron = util.CallerCron

// Fields defined here so drivers can handle generic order/filter arguments.
const (
	FieldCreatedAt    = "created_at"
	FieldWorkflowPath = "workflow_path"
	FieldInvoker      = "invoker"
	FieldStatus       = "status" // The driver is responsible for converting string to enum, not the caller.
)

// Types of filters defined here. Not all types of filters are supported for all fields.
const (
	FilterKindPrefix   = "prefix"
	FilterKindContains = "contains"
	FilterKindMatch    = "match"
	FilterKindAfter    = "after"
	FilterKindBefore   = "before"
)

// Order defines a generic way to apply optional ordering to a list query.
type Order struct {
	Field      string
	Descending bool
}

// Filter defines a generic way to apply optional filtering to a list query.
type Filter struct {
	Field string
	Kind  string
	Value interface{}
}

// ListOpts defines a generic way to apply common optional modifiers to list requests.
type ListOpts struct {
	Limit   int
	Offset  int
	Orders  []Order
	Filters []Filter
}

// InstanceData is the struct that matches the instance data table.
type InstanceData struct {
	ID             uuid.UUID
	NamespaceID    uuid.UUID
	Namespace      string
	RootInstanceID uuid.UUID
	Server         uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
	EndedAt        *time.Time
	Deadline       *time.Time
	Status         InstanceStatus
	WorkflowPath   string
	ErrorCode      string
	Invoker        string
	Definition     []byte
	DescentInfo    []byte
	TelemetryInfo  []byte
	RuntimeInfo    []byte
	ChildrenInfo   []byte
	Input          []byte
	LiveData       []byte
	StateMemory    []byte
	Output         []byte
	ErrorMessage   []byte
	Metadata       []byte
}

// GetNamespaceInstancesResults returns the results as well as the total number that would be returned if LIMIT & OFFSET were both zero.
type GetNamespaceInstancesResults struct {
	Total   int
	Results []InstanceData
}

// CreateInstanceDataArgs defines the required arguments for creating a new instance data record.
type CreateInstanceDataArgs struct {
	ID             uuid.UUID
	NamespaceID    uuid.UUID
	Namespace      string
	RootInstanceID uuid.UUID
	Server         uuid.UUID
	Invoker        string
	WorkflowPath   string
	Definition     []byte
	Input          []byte
	LiveData       []byte
	TelemetryInfo  []byte
	DescentInfo    []byte
	RuntimeInfo    []byte
	ChildrenInfo   []byte
}

// UpdateInstanceDataArgs defines the possible arguments for updating an existing instance data record.
type UpdateInstanceDataArgs struct {
	BypassOwnershipCheck bool            `json:"bypass_ownership_check"`
	Server               uuid.UUID       `json:"server"`
	EndedAt              *time.Time      `json:"ended_at,omitempty"`
	Deadline             *time.Time      `json:"deadline,omitempty"`
	Status               *InstanceStatus `json:"status,omitempty"`
	ErrorCode            *string         `json:"error_code,omitempty"`
	TelemetryInfo        *[]byte         `json:"telemetry_info,omitempty"`
	RuntimeInfo          *[]byte         `json:"runtime_info,omitempty"`
	ChildrenInfo         *[]byte         `json:"children_info,omitempty"`
	LiveData             *[]byte         `json:"live_data,omitempty"`
	StateMemory          *[]byte         `json:"state_memory,omitempty"`
	Output               *[]byte         `json:"output,omitempty"`
	ErrorMessage         *[]byte         `json:"error_message,omitempty"`
	Metadata             *[]byte         `json:"metadata,omitempty"`
}

// InstanceMessageData is the struct that matches the instance messages table.
type InstanceMessageData struct {
	ID         uuid.UUID
	InstanceID uuid.UUID
	CreatedAt  time.Time
	Payload    []byte
}

// EnqueueInstanceMessageArgs defines the required arguments for enqueueing an instance message.
type EnqueueInstanceMessageArgs struct {
	InstanceID uuid.UUID
	Payload    []byte
}

// InstanceCounts defined the return object for the metrics function.
type InstanceCounts struct {
	Complete  int
	Failed    int
	Crashed   int
	Cancelled int
	Pending   int
	Total     int
}

type InstanceDataQuery interface {
	// UpdateInstanceData updates the instance record. It only applies non-nil arguments. It returns the updated record.
	UpdateInstanceData(ctx context.Context, args *UpdateInstanceDataArgs) error

	// GetMost returns almost all fields, excluding only one or two fields that the engine is unlikely to need (input, output & metadata)
	GetMost(ctx context.Context) (*InstanceData, error)

	// GetSummary returns all fields that should be reasonably small, to avoid potentially loading megabytes of data unnecessarily.
	GetSummary(ctx context.Context) (*InstanceData, error)

	// GetSummaryWithInput returns everything GetSummary does, as well as the input field.
	GetSummaryWithInput(ctx context.Context) (*InstanceData, error)

	// GetSummaryWithOutput returns everything GetSummary does, as well as the output field.
	GetSummaryWithOutput(ctx context.Context) (*InstanceData, error)

	// GetSummaryWithMetadata returns everything GetSummary does, as well as the metadata field.
	GetSummaryWithMetadata(ctx context.Context) (*InstanceData, error)

	// EnqueueMessage adds a message to the instance's message queue.
	EnqueueMessage(ctx context.Context, args *EnqueueInstanceMessageArgs) error

	// PopMessage returns the most recent message out of the queue.
	PopMessage(ctx context.Context) (*InstanceMessageData, error)
}

type Store interface {
	// ForInstanceID creates an InstanceDataQuery object, from which queries related to a specific instance can be created.
	ForInstanceID(id uuid.UUID) InstanceDataQuery

	// CreateInstanceData creates a new row in the database.
	// NOTE: the created_at and updated_at fields returned are incorrect. Correcting them would require an additional
	// 		SQL query, and the performance tradeoff is not worth it.
	CreateInstanceData(ctx context.Context, args *CreateInstanceDataArgs) (*InstanceData, error)

	// GetNamespaceInstances returns a list of instances associated with the given namespace ID.
	// Unless overwritten with list options, the default ordering should be by created_at desc with no filters, no limit and no offset.
	GetNamespaceInstances(ctx context.Context, nsID uuid.UUID, opts *ListOpts) (*GetNamespaceInstancesResults, error)

	// GetHangingInstances returns a list of all instances where deadline has been exceeded and status is unfinished
	GetHangingInstances(ctx context.Context) ([]InstanceData, error)

	// GetHomelessInstances returns a list of all unfinished instances where updated_at hasn't been touched in a while, so that the engine can attempt to auto-resume work on them
	GetHomelessInstances(ctx context.Context, t time.Time) ([]InstanceData, error)

	// DeleteOldInstances deletes all instances that have terminated and end_at a long time ago
	DeleteOldInstances(ctx context.Context, before time.Time) error

	// AssertNoParallelCron attempts to detect if another machine in a HA environment may have already triggered an instance that we're just about to create ourselves.
	// It does this by checking if a record of an instance was created within the last 30s for the given workflow ID.
	AssertNoParallelCron(ctx context.Context, nsID uuid.UUID, wfPath string) error

	// GetNamespaceInstanceCounts returns some instance metrics.
	GetNamespaceInstanceCounts(ctx context.Context, nsID uuid.UUID, wfPath string) (*InstanceCounts, error)
}

type (
	InstanceCanceller func(ctx context.Context, namespace, instanceID string) error
	InstanceStarter   func(ctx context.Context, namespace, path string, input []byte) (*InstanceData, error)
)

type InstanceManager struct {
	Cancel InstanceCanceller
	Start  InstanceStarter
}
