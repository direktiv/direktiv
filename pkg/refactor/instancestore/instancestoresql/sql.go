package instancestoresql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	table               = "instances_v2"
	fieldID             = "id"
	fieldNamespaceID    = "namespace_id"
	fieldWorkflowID     = "workflow_id"
	fieldRevisionID     = "revision_id"
	fieldRootInstanceID = "root_instance_id"
	fieldCreatedAt      = "created_at"
	fieldUpdatedAt      = "updated_at"
	fieldEndedAt        = "ended_at"
	fieldDeadline       = "deadline"
	fieldStatus         = "status"
	fieldCalledAs       = "called_as"
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

	desc = "desc"
)

var (
	mostFields = []string{
		fieldID, fieldNamespaceID, fieldWorkflowID, fieldRevisionID, fieldRootInstanceID,
		fieldCreatedAt, fieldUpdatedAt, fieldEndedAt, fieldDeadline, fieldStatus, fieldCalledAs,
		fieldErrorCode, fieldInvoker, fieldDefinition, fieldSettings, fieldDescentInfo, fieldTelemetryInfo,
		fieldRuntimeInfo, fieldChildrenInfo /*"input",*/, fieldLiveData, fieldStateMemory, fieldErrorMessage,
	}

	summaryFields = []string{
		fieldID, fieldNamespaceID, fieldWorkflowID, fieldRevisionID, fieldRootInstanceID,
		fieldCreatedAt, fieldUpdatedAt, fieldEndedAt, fieldDeadline, fieldStatus, fieldCalledAs,
		fieldErrorCode, fieldInvoker, fieldErrorMessage,
	}
)

type sqlInstanceStore struct {
	db     *gorm.DB
	logger *zap.SugaredLogger
}

func (s *sqlInstanceStore) ForInstanceID(id uuid.UUID) instancestore.InstanceDataQuery {
	return &instanceDataQuery{
		instanceID: id,
		db:         s.db,
		logger:     s.logger,
	}
}

var _ instancestore.Store = &sqlInstanceStore{} // Ensures sqlInstanceStore struct conforms to instanceStore.Store interface.

func NewSQLInstanceStore(db *gorm.DB, logger *zap.SugaredLogger) instancestore.Store {
	return &sqlInstanceStore{
		db:     db,
		logger: logger,
	}
}

func (s *sqlInstanceStore) CreateInstanceData(ctx context.Context, args *instancestore.CreateInstanceDataArgs) (*instancestore.InstanceData, error) {
	idata := &instancestore.InstanceData{
		ID:             args.ID,
		NamespaceID:    args.NamespaceID,
		WorkflowID:     args.WorkflowID,
		RevisionID:     args.RevisionID,
		RootInstanceID: args.RootInstanceID,
		Status:         instancestore.InstanceStatusPending,
		CalledAs:       args.CalledAs,
		ErrorCode:      "",
		Invoker:        args.Invoker,
		Definition:     args.Definition,
		Settings:       args.Settings,
		DescentInfo:    args.DescentInfo,
		TelemetryInfo:  args.TelemetryInfo,
		RuntimeInfo:    []byte(`{}`),
		ChildrenInfo:   []byte(`{}`),
		Input:          args.Input,
		LiveData:       []byte(`{}`),
		StateMemory:    []byte(``),
		Output:         nil,
		ErrorMessage:   nil,
		Metadata:       nil,
	}

	columns := []string{
		fieldID, fieldNamespaceID, fieldWorkflowID, fieldRevisionID, fieldRootInstanceID,
		fieldStatus, fieldCalledAs, fieldErrorCode, fieldInvoker, fieldDefinition,
		fieldSettings, fieldDescentInfo, fieldTelemetryInfo, fieldRuntimeInfo,
		fieldChildrenInfo, fieldInput, fieldLiveData, fieldStateMemory,
	}
	into := strings.Join(columns, ", ")
	valPlaceholders := strings.Repeat("?, ", len(columns)-1) + "?"
	query := fmt.Sprintf(`INSERT INTO %s(%s) VALUES (%s)`, table, into, valPlaceholders)

	s.logger.Debug("CreateInstanceData executing SQL query: %s", query)

	res := s.db.WithContext(ctx).Exec(query,
		idata.ID, idata.NamespaceID, idata.WorkflowID, idata.RevisionID, idata.RootInstanceID,
		idata.Status, idata.CalledAs, idata.ErrorCode, idata.Invoker, idata.Definition,
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

func generateGetNamespaceInstancesOrderings(opts *instancestore.ListOpts) (string, error) {
	if len(opts.Orders) == 0 {
		return " " + fieldCreatedAt + " " + desc, nil
	} else {
		var orderStrings []string
		for _, order := range opts.Orders {
			var s string
			switch order.Field {
			case instancestore.FieldCreatedAt:
				s = fieldCreatedAt
			default:
				return "", fmt.Errorf("order field '%s': %w", order.Field, instancestore.ErrBadListOpts)
			}

			if order.Descending {
				s += " " + desc
			}

			orderStrings = append(orderStrings, s)
		}
		return ` ORDER BY ` + strings.Join(orderStrings, ", "), nil
	}
}

func generateGetNamespaceInstancesFilters(opts *instancestore.ListOpts) ([]string, []interface{}, error) {
	var clauses []string
	var vals []interface{}
	for idx := range opts.Filters {
		filter := opts.Filters[idx]
		var clause string
		var val interface{}
		switch filter.Field {
		case fieldNamespaceID:
			if filter.Kind == instancestore.FilterKindMatch {
				clause = fieldNamespaceID + " = ?"
				val = filter.Value
			} else {
				return nil, nil, fmt.Errorf("filter kind '%s' for use with field '%s': %w", filter.Kind, filter.Field, instancestore.ErrBadListOpts)
			}

		case instancestore.FieldCreatedAt:
			if filter.Kind == instancestore.FilterKindBefore {
				clause = fieldCreatedAt + " < ?"
				val = filter.Value
			} else if filter.Kind == instancestore.FilterKindAfter {
				clause = fieldCreatedAt + " > ?"
				val = filter.Value
			} else {
				return nil, nil, fmt.Errorf("filter kind '%s' for use with field '%s': %w", filter.Kind, filter.Field, instancestore.ErrBadListOpts)
			}

		case fieldDeadline:
			if filter.Kind == instancestore.FilterKindBefore {
				clause = fieldDeadline + " < ?"
				val = filter.Value
			} else if filter.Kind == instancestore.FilterKindAfter {
				clause = fieldDeadline + " > ?"
				val = filter.Value
			} else {
				return nil, nil, fmt.Errorf("filter kind '%s' for use with field '%s': %w", filter.Kind, filter.Field, instancestore.ErrBadListOpts)
			}

		case instancestore.FieldCalledAs:
			if filter.Kind == instancestore.FilterKindPrefix {
				clause = fieldCalledAs + " LIKE ?"
				val = filter.Value.(string) + "%"
			} else if filter.Kind == instancestore.FilterKindContains {
				clause = fieldCalledAs + " LIKE ?"
				val = "%" + filter.Value.(string) + "%"
			} else {
				return nil, nil, fmt.Errorf("filter kind '%s' for use with field '%s': %w", filter.Kind, filter.Field, instancestore.ErrBadListOpts)
			}

		case instancestore.FieldStatus:
			if filter.Kind == instancestore.FilterKindMatch {
				clause = fieldStatus + " = ?"
				val = filter.Value
			} else if filter.Kind == "<" {
				clause = fieldStatus + " < ?"
				val = filter.Value
			} else {
				return nil, nil, fmt.Errorf("filter kind '%s' for use with field '%s': %w", filter.Kind, filter.Field, instancestore.ErrBadListOpts)
			}

		case instancestore.FieldInvoker:
			if filter.Kind == instancestore.FilterKindMatch {
				clause = fieldInvoker + " LIKE ?"
				val = filter.Value.(string)
			} else if filter.Kind == instancestore.FilterKindContains {
				clause = fieldInvoker + " LIKE ?"
				val = "%" + filter.Value.(string) + "%"
			} else {
				return nil, nil, fmt.Errorf("filter kind '%s' for use with field '%s': %w", filter.Kind, filter.Field, instancestore.ErrBadListOpts)
			}

		default:
			return nil, nil, fmt.Errorf("filter field '%s': %w", filter.Field, instancestore.ErrBadListOpts)
		}

		clauses = append(clauses, clause)
		vals = append(vals, val)
	}

	return clauses, vals, nil
}

func wheres(clauses ...string) string {
	if len(clauses) == 0 {
		return ""
	}
	if len(clauses) == 1 {
		return ` WHERE ` + clauses[0]
	}
	return ` WHERE (` + strings.Join(clauses, ") AND (") + `)`
}

func (s *sqlInstanceStore) generateGetInstancesQuery(ctx context.Context, columns []string, opts *instancestore.ListOpts) ([]instancestore.InstanceData, int, error) {

	clauses, vals, err := generateGetNamespaceInstancesFilters(opts)
	if err != nil {
		return nil, 0, err
	}

	orderings, err := generateGetNamespaceInstancesOrderings(opts)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`SELECT %s FROM %s`, strings.Join(columns, ", "), table)
	query += wheres(clauses...)
	query += orderings

	if opts.Limit > 0 {
		query += fmt.Sprintf(` LIMIT %d`, opts.Limit)

		if opts.Offset > 0 {
			query += fmt.Sprintf(` OFFSET %d`, opts.Offset)
		}
	}

	s.logger.Debug("generateGetInstancesQuery executing SQL query: %s", query)

	var idatas []instancestore.InstanceData

	res := s.db.WithContext(ctx).Raw(query, vals...).Find(&idatas)
	if res.Error != nil {
		return nil, 0, res.Error
	}

	return idatas, int(res.RowsAffected), nil

}

func (s *sqlInstanceStore) GetNamespaceInstances(ctx context.Context, nsID uuid.UUID, opts *instancestore.ListOpts) (*instancestore.GetNamespaceInstancesResults, error) {
	opts.Filters = append([]instancestore.Filter{{
		Field: fieldNamespaceID,
		Kind:  instancestore.FilterKindMatch,
		Value: nsID,
	}}, opts.Filters...)

	idatas, k, err := s.generateGetInstancesQuery(ctx, summaryFields, opts)
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
		Value: instancestore.InstanceStatusComplete.String(),
	}, {
		Field: fieldDeadline,
		Kind:  instancestore.FilterKindBefore,
		Value: time.Now(),
	}}, opts.Filters...)

	idatas, _, err := s.generateGetInstancesQuery(ctx, summaryFields, opts)
	if err != nil {
		return nil, err
	}

	return idatas, nil
}

func (s *sqlInstanceStore) DeleteOldInstances(ctx context.Context, before time.Time) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE %s >= ? AND %s < ?`, table, fieldStatus, fieldEndedAt)
	s.logger.Debug("DeleteOldInstances executing SQL query: %s", query)

	res := s.db.WithContext(ctx).Exec(
		query,
		instancestore.InstanceStatusComplete, before,
	)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (s *sqlInstanceStore) AssertNoParallelCron(ctx context.Context, wfID uuid.UUID) error {
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE %s = ? AND %s = ? AND %s > ?`, table, fieldInvoker, fieldWorkflowID, fieldCreatedAt)
	s.logger.Debug("AssertNoParallelCron executing SQL query: %s", query)

	var k int64
	res := s.db.WithContext(ctx).Raw(
		query,
		instancestore.InvokerCron, wfID, time.Now().Add(time.Second*30),
	).First(&k)
	if res.Error != nil {
		return res.Error
	}

	if k != 0 {
		return instancestore.ErrParallelCron
	}

	return nil
}
