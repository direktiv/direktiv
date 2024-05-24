package instancestoresql

import (
	"context"
	"errors"
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
	fieldSettings       = "settings" // TODO: alan, remove this when we're ready to make a bundle of breaking database changes
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
		fieldErrorCode, fieldInvoker, fieldDefinition, fieldDescentInfo, fieldTelemetryInfo,
		fieldRuntimeInfo, fieldChildrenInfo, fieldLiveData, fieldStateMemory, fieldErrorMessage,
	}

	summaryFields = []string{
		fieldID, fieldNamespaceID, fieldNamespace, fieldRootInstanceID, fieldServer,
		fieldCreatedAt, fieldUpdatedAt, fieldEndedAt, fieldDeadline, fieldStatus, fieldWorkflowPath,
		fieldErrorCode, fieldInvoker, fieldDefinition, fieldDescentInfo, fieldTelemetryInfo,
		fieldRuntimeInfo, fieldChildrenInfo, fieldErrorMessage,
		`length(` + fieldInput + `) as input_length`, `length(` + fieldOutput + `) as output_length`, `length(` + fieldMetadata + `) as metadata_length`,
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
		make([]byte, 0), idata.DescentInfo, idata.TelemetryInfo, idata.RuntimeInfo,
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

func (s *sqlInstanceStore) GetNamespaceInstanceCounts(ctx context.Context, nsID uuid.UUID, wfPath string) (*instancestore.InstanceCounts, error) {
	query := fmt.Sprintf(`SELECT COUNT(%s), %s FROM %s WHERE %s = ? AND %s = ? GROUP BY %s`, fieldID, fieldStatus, table, fieldNamespaceID, fieldWorkflowPath, fieldStatus)

	x := make([]map[string]interface{}, 0)
	res := s.db.WithContext(ctx).Raw(
		query,
		nsID, wfPath,
	).Find(&x)
	if res.Error != nil {
		return nil, res.Error
	}

	m := make(map[instancestore.InstanceStatus]int)

	var total int

	for _, y := range x {
		// NOTE: it seems that the sqlite driver and the pq drivers name these values differently and store them as different type,
		// so we have to try two different options. There has got to be a better way to do this...
		var status int
		v := y["status"]
		if k1, ok := v.(int32); ok {
			status = int(k1)
		} else if k2, ok := v.(int64); ok {
			status = int(k2)
		}

		var count int
		if v, exists := y["count"]; exists {
			count = int(v.(int64)) //nolint
		} else if v, exists = y["COUNT(id)"]; exists {
			count = int(v.(int64)) //nolint
		} else {
			return nil, errors.New("unexpected database response")
		}

		m[instancestore.InstanceStatus(status)] = count
		total += count
	}

	return &instancestore.InstanceCounts{
		Complete:  m[instancestore.InstanceStatusComplete],
		Failed:    m[instancestore.InstanceStatusFailed],
		Crashed:   m[instancestore.InstanceStatusCrashed],
		Cancelled: m[instancestore.InstanceStatusCancelled],
		Pending:   m[instancestore.InstanceStatusPending],
		Total:     total,
	}, nil
}
