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
	ID             uuid.UUID
	NamespaceID    uuid.UUID
	WorkflowID     uuid.UUID
	RevisionID     uuid.UUID
	RootInstanceID uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
	EndedAt        *time.Time
	Deadline       *time.Time
	Status         string // TODO: alan, turn this into an enum so we can do arithmetic comparisons on it
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
	GetNamespaceInstances(ctx context.Context, nsID uuid.UUID, opts *GetInstancesListOpts) ([]*InstanceData, error)

	// TODO: alan, GetHangingInstances :: query to get a list of all instances where deadline has been exceeded and status is unfinished
	// TODO: alan, DeleteOldInstances :: query to delete all instances that have terminated and end_at a long time ago
	// TODO: alan, CheckForParallelCron :: query to try and detect if another machine in a HA environment may have already triggered an instance that we're just about to create ourselves
}

// TODO: alan, discuss all that follows with Yassir to brainstorm nicest way to handle complex queries
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
