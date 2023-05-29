package instancestore

import (
	"context"
	"errors"
	"time"

	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrParallelCron = errors.New("a parallel cron already exists")
)

type InstanceStatus int

const (
	InstanceStatusUndefined = iota
	InstanceStatusPending
	InstanceStatusComplete
	InstanceStatusFailed
	InstanceStatusCrashed
)

var (
	instanceStatusStrings = []string{"undefined", util.InstanceStatusPending, util.InstanceStatusComplete, util.InstanceStatusFailed, util.InstanceStatusCrashed}
)

func (status InstanceStatus) String() string {
	return instanceStatusStrings[status]
}

const (
	InvokerCron = util.CallerCron
)

type InstanceData struct {
	ID             uuid.UUID
	NamespaceID    uuid.UUID
	WorkflowID     uuid.UUID
	RevisionID     uuid.UUID
	RootInstanceID uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
	EndedAt        *time.Time
	Deadline       *time.Time
	Status         InstanceStatus
	CalledAs       string
	ErrorCode      string
	Invoker        string
	Definition     []byte // TODO: alan, we should strip comments
	Settings       []byte
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
	EndedAt       *time.Time
	Deadline      *time.Time
	Status        *InstanceStatus
	ErrorCode     *string
	TelemetryInfo *[]byte
	RuntimeInfo   *[]byte
	ChildrenInfo  *[]byte
	LiveData      *[]byte
	StateMemory   *[]byte
	Output        *[]byte
	ErrorMessage  *[]byte
	Metadata      *[]byte
	// TODO: alan, should we consider compressing these binary fields?
}

type InstanceDataQuery interface {
	// UpdateInstanceData updates the instance record. It only applies non-nil arguments.
	UpdateInstanceData(ctx context.Context, args *UpdateInstanceDataArgs) (*InstanceData, error)

	// TODO: alan, implement more fine-grained setters?
	// - UpdateInstanceTransition
	// - UpdateInstanceYield
	// - UpdateInstanceTerminateTerminate

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
}

type Store interface {
	// TODO: alan, discuss with Yassir if the namespace ID should be an argument here
	// ForInstanceID creates an InstanceDataQuery object, from which queries related to a specific instance can be created.
	ForInstanceID(id uuid.UUID) InstanceDataQuery

	// CreateInstanceData creates a new row in the database.
	CreateInstanceData(ctx context.Context, args *CreateInstanceDataArgs) (*InstanceData, error)

	// TODO: alan, pagination/filtering on this list
	// GetNamespaceInstances returns a list of instances associated with the given namespace ID.
	GetNamespaceInstances(ctx context.Context, nsID uuid.UUID, opts *GetInstancesListOpts) (*GetNamespaceInstancesResults, error)

	// GetHangingInstances returns a list of all instances where deadline has been exceeded and status is unfinished
	GetHangingInstances(ctx context.Context) ([]*InstanceData, error)

	// DeleteOldInstances deletes all instances that have terminated and end_at a long time ago
	DeleteOldInstances(ctx context.Context, before time.Time) error

	// AssertNoParallelCron attempts to detect if another machine in a HA environment may have already triggered an instance that we're just about to create ourselves.
	// It does this by checking if a record of an instance was created within the last 30s for the given workflow ID.
	AssertNoParallelCron(ctx context.Context, wfID uuid.UUID) error
}

// TODO: alan, discuss all that follows with Yassir to brainstorm nicest way to handle complex queries
type GetNamespaceInstancesResults struct {
	Total   int
	Results []*InstanceData
}

type Order struct {
	Field      string
	Descending bool
}

type Filter struct {
	Field string
	Kind  string
	Value interface{}
}

type GetInstancesListOpts struct {
	LimitResults  int
	OffsetResults int
	Order         []Order
	Filter        []Filter
}

const (
	FieldCreatedAt = "created_at"
	FieldCalledAs  = "called_as"
	FieldStatus    = "status"
	FieldInvoker   = "invoker"
)

const (
	FilterKindPrefix   = "prefix"
	FilterKindContains = "contains"
	FilterKindMatch    = "match"
	FilterKindAfter    = "after"
	FilterKindBefore   = "before"
)

func (opts *GetInstancesListOpts) Limit(limit int) *GetInstancesListOpts {
	opts.LimitResults = limit
	return opts
}

func (opts *GetInstancesListOpts) Offset(offset int) *GetInstancesListOpts {
	opts.OffsetResults = offset
	return opts
}

func (opts *GetInstancesListOpts) OrderByCreatedAtDesc() *GetInstancesListOpts {
	opts.Order = append(opts.Order, Order{
		Field:      FieldCreatedAt,
		Descending: true,
	})
	return opts
}

func (opts *GetInstancesListOpts) FilterPrefixCalledAs(prefix string) *GetInstancesListOpts {
	opts.Filter = append(opts.Filter, Filter{
		Field: FieldCalledAs,
		Kind:  FilterKindPrefix,
		Value: prefix,
	})
	return opts
}

func (opts *GetInstancesListOpts) FilterContainsCalledAs(substr string) *GetInstancesListOpts {
	opts.Filter = append(opts.Filter, Filter{
		Field: FieldCalledAs,
		Kind:  FilterKindContains,
		Value: substr,
	})
	return opts
}

func (opts *GetInstancesListOpts) FilterBeforeCreatedAs(t time.Time) *GetInstancesListOpts {
	opts.Filter = append(opts.Filter, Filter{
		Field: FieldCreatedAt,
		Kind:  FilterKindBefore,
		Value: t,
	})
	return opts
}

func (opts *GetInstancesListOpts) FilterAfterCreatedAs(t time.Time) *GetInstancesListOpts {
	opts.Filter = append(opts.Filter, Filter{
		Field: FieldCreatedAt,
		Kind:  FilterKindAfter,
		Value: t,
	})
	return opts
}

func (opts *GetInstancesListOpts) FilterMatchStatus(status string) *GetInstancesListOpts {
	opts.Filter = append(opts.Filter, Filter{
		Field: FieldStatus,
		Kind:  FilterKindMatch,
		Value: status,
	})
	return opts
}

func (opts *GetInstancesListOpts) FilterContainsStatus(substr string) *GetInstancesListOpts {
	opts.Filter = append(opts.Filter, Filter{
		Field: FieldStatus,
		Kind:  FilterKindContains,
		Value: substr,
	})
	return opts
}

func (opts *GetInstancesListOpts) FilterMatchInvoker(invoker string) *GetInstancesListOpts {
	opts.Filter = append(opts.Filter, Filter{
		Field: FieldInvoker,
		Kind:  FilterKindMatch,
		Value: invoker,
	})
	return opts
}

func (opts *GetInstancesListOpts) FilterContainsInvoker(substr string) *GetInstancesListOpts {
	opts.Filter = append(opts.Filter, Filter{
		Field: FieldInvoker,
		Kind:  FilterKindContains,
		Value: substr,
	})
	return opts
}
