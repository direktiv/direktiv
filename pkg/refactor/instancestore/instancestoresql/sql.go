package instancestoresql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	table               = "instances_v2"
	fieldID             = "id"
	fieldNamespaceID    = "namespace_id"
	fieldNamespace      = "namespace"
	fieldRootInstanceID = "root_instance_id"
	fieldServer         = "server"
	fieldCreatedAt      = "created_at"
	fieldUpdatedAt      = "updated_at"
	fieldEndedAt        = "ended_at"
	fieldDeadline       = "deadline"
	fieldStatus         = "status"
	fieldWorkflowPath   = "workflow_path"
	fieldErrorCode      = "error_code"
	fieldInvoker        = "invoker"
	fieldDefinition     = "definition"
	fieldSettings       = "settings"
	fieldDescentInfo    = "descent_info"
	fieldTelemetryInfo  = "telemetry_info"
	fieldRuntimeInfo    = "runtime_info"
	fieldChildrenInfo   = "children_info"
	fieldInput          = "input"
	fieldLiveData       = "live_data"
	fieldStateMemory    = "state_memory"
	fieldOutput         = "output"
	fieldErrorMessage   = "error_message"
	fieldMetadata       = "metadata"

	messagesTable                  = "instance_messages"
	fieldInstanceMessageID         = "id"
	fieldInstanceMessageInstanceID = "instance_id"
	fieldInstanceMessageCreatedAt  = "created_at"
	fieldInstanceMessagePayload    = "payload"

	desc = "desc"
)

var (
	mostFields = []string{
		fieldID, fieldNamespaceID, fieldNamespace, fieldRootInstanceID, fieldServer,
		fieldCreatedAt, fieldUpdatedAt, fieldEndedAt, fieldDeadline, fieldStatus, fieldWorkflowPath,
		fieldErrorCode, fieldInvoker, fieldDefinition, fieldSettings, fieldDescentInfo, fieldTelemetryInfo,
		fieldRuntimeInfo, fieldChildrenInfo, fieldLiveData, fieldStateMemory, fieldErrorMessage,
	}

	summaryFields = []string{
		fieldID, fieldNamespaceID, fieldNamespace, fieldRootInstanceID, fieldServer,
		fieldCreatedAt, fieldUpdatedAt, fieldEndedAt, fieldDeadline, fieldStatus, fieldWorkflowPath,
		fieldErrorCode, fieldInvoker, fieldSettings, fieldDescentInfo, fieldTelemetryInfo,
		fieldRuntimeInfo, fieldChildrenInfo, fieldErrorMessage,
	}
)

type sqlInstanceStore struct {
	db *gorm.DB
}

func (s *sqlInstanceStore) ForInstanceID(id uuid.UUID) instancestore.InstanceDataQuery {
	return &instanceDataQuery{
		instanceID: id,
		db:         s.db,
	}
}

var _ instancestore.Store = &sqlInstanceStore{} // Ensures sqlInstanceStore struct conforms to instanceStore.Store interface.

func NewSQLInstanceStore(db *gorm.DB) instancestore.Store {
	return &sqlInstanceStore{
		db: db,
	}
}

func (s *sqlInstanceStore) CreateInstanceData(ctx context.Context, args *instancestore.CreateInstanceDataArgs) (*instancestore.InstanceData, error) {
	idata := &instancestore.InstanceData{
		ID:             args.ID,
		NamespaceID:    args.NamespaceID,
		Namespace:      args.Namespace,
		RootInstanceID: args.RootInstanceID,
		Server:         args.Server,
		Status:         instancestore.InstanceStatusPending,
		WorkflowPath:   args.WorkflowPath,
		ErrorCode:      "",
		Invoker:        args.Invoker,
		Definition:     args.Definition,
		Settings:       args.Settings,
		DescentInfo:    args.DescentInfo,
		TelemetryInfo:  args.TelemetryInfo,
		RuntimeInfo:    args.RuntimeInfo,
		ChildrenInfo:   args.ChildrenInfo,
		Input:          args.Input,
		LiveData:       args.LiveData,
		StateMemory:    []byte(`{}`),
		Output:         nil,
		ErrorMessage:   nil,
		Metadata:       nil,
	}

	columns := []string{
		fieldID, fieldNamespaceID, fieldNamespace, fieldRootInstanceID, fieldServer,
		fieldStatus, fieldWorkflowPath, fieldErrorCode, fieldInvoker, fieldDefinition,
		fieldSettings, fieldDescentInfo, fieldTelemetryInfo, fieldRuntimeInfo,
		fieldChildrenInfo, fieldInput, fieldLiveData, fieldStateMemory,
	}
	query := generateInsertQuery(table, columns)

	res := s.db.WithContext(ctx).Exec(query,
		idata.ID, idata.NamespaceID, idata.Namespace, idata.RootInstanceID, idata.Server,
		idata.Status, idata.WorkflowPath, idata.ErrorCode, idata.Invoker, idata.Definition,
		idata.Settings, idata.DescentInfo, idata.TelemetryInfo, idata.RuntimeInfo,
		idata.ChildrenInfo, idata.Input, idata.LiveData, idata.StateMemory)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpected gorm create count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return idata, nil
}

func (s *sqlInstanceStore) performGetInstancesQuery(ctx context.Context, columns []string, opts *instancestore.ListOpts) ([]instancestore.InstanceData, int, error) {
	countQuery, query, vals, err := generateGetInstancesQueries(columns, opts)
	if err != nil {
		return nil, 0, err
	}

	var count int
	var idatas []instancestore.InstanceData

	res := s.db.WithContext(ctx).Raw(query, vals...).Find(&idatas)
	if res.Error != nil {
		return nil, 0, res.Error
	}

	count = len(idatas)

	if opts != nil && opts.Limit != 0 {
		res = s.db.WithContext(ctx).Raw(countQuery, vals...).First(&count)
		if res.Error != nil {
			return nil, 0, res.Error
		}
	}

	return idatas, count, nil
}

func (s *sqlInstanceStore) GetNamespaceInstances(ctx context.Context, nsID uuid.UUID, opts *instancestore.ListOpts) (*instancestore.GetNamespaceInstancesResults, error) {
	if opts == nil {
		opts = &instancestore.ListOpts{}
	}

	opts.Filters = append([]instancestore.Filter{{
		Field: fieldNamespaceID,
		Kind:  instancestore.FilterKindMatch,
		Value: nsID,
	}}, opts.Filters...)

	idatas, k, err := s.performGetInstancesQuery(ctx, summaryFields, opts)
	if err != nil {
		return nil, err
	}

	return &instancestore.GetNamespaceInstancesResults{
		Total:   k,
		Results: idatas,
	}, nil
}

func (s *sqlInstanceStore) GetHangingInstances(ctx context.Context) ([]instancestore.InstanceData, error) {
	opts := new(instancestore.ListOpts)
	opts.Filters = append([]instancestore.Filter{{
		Field: fieldStatus,
		Kind:  "<",
		Value: instancestore.InstanceStatusComplete,
	}, {
		Field: fieldDeadline,
		Kind:  instancestore.FilterKindBefore,
		Value: time.Now().UTC(),
	}}, opts.Filters...)

	idatas, _, err := s.performGetInstancesQuery(ctx, summaryFields, opts)
	if err != nil {
		return nil, err
	}

	return idatas, nil
}

func (s *sqlInstanceStore) GetHomelessInstances(ctx context.Context, t time.Time) ([]instancestore.InstanceData, error) {
	query := fmt.Sprintf(`
SELECT DISTINCT {table0}.%s, {table0}.%s, {table0}.%s
FROM {table0}
INNER JOIN {table1} ON {table0}.%s={table1}.%s
WHERE {table0}.%s < ? AND {table0}.%s < ?
`, fieldID, fieldServer, fieldUpdatedAt, fieldID, fieldInstanceMessageInstanceID, fieldStatus, fieldUpdatedAt)
	query = strings.ReplaceAll(query, "{table0}", table)
	query = strings.ReplaceAll(query, "{table1}", messagesTable)

	var idatas []instancestore.InstanceData

	res := s.db.WithContext(ctx).Raw(query, instancestore.InstanceStatusComplete, t).Find(&idatas)
	if res.Error != nil {
		return nil, res.Error
	}

	return idatas, nil
}

func (s *sqlInstanceStore) DeleteOldInstances(ctx context.Context, before time.Time) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE %s >= ? AND %s < ?`, table, fieldStatus, fieldEndedAt)

	res := s.db.WithContext(ctx).Exec(
		query,
		instancestore.InstanceStatusComplete, before.UTC(),
	)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (s *sqlInstanceStore) AssertNoParallelCron(ctx context.Context, wfPath string) error {
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE %s = ? AND %s = ? AND %s > ?`, table, fieldInvoker, fieldWorkflowPath, fieldCreatedAt)

	var k int64
	res := s.db.WithContext(ctx).Raw(
		query,
		instancestore.InvokerCron, wfPath, time.Now().UTC().Add(-30*time.Second),
	).First(&k)
	if res.Error != nil {
		return res.Error
	}

	if k != 0 {
		return instancestore.ErrParallelCron
	}

	return nil
}
