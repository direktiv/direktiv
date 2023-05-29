package instancestoresql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	table            = "instances_v2"
	fieldCreatedAt   = "created_at"
	fieldEndedAt     = "ended_at"
	fieldCalledAs    = "called_as"
	fieldStatus      = "status"
	fieldInvoker     = "invoker"
	fieldDeadline    = "deadline"
	fieldNamespaceID = "namespace_id"
	fieldWorkflowID  = "workflow_id"
	desc             = "desc"
)

var (
	summaryFields = []string{
		"id", fieldNamespaceID, fieldWorkflowID, "revision_id", "root_instance_id",
		fieldCreatedAt, "updated_at", fieldEndedAt, fieldDeadline, fieldStatus, fieldCalledAs,
		"error_code", fieldInvoker,
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
	res := s.db.WithContext(ctx).Table(table).Create(idata)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpected gorm create count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return idata, nil
}

func applyGetNamespaceInstancesOrderings(query *gorm.DB, opts *instancestore.GetInstancesListOpts) *gorm.DB {
	if len(opts.Order) == 0 {
		query = query.Order(fieldCreatedAt + " " + desc)
	} else {
		for _, order := range opts.Order {
			var s string
			switch order.Field {
			case instancestore.FieldCreatedAt:
				s = fieldCreatedAt
			default:
				panic(fmt.Errorf("unexpected order field '%s'", order.Field))
			}

			if order.Descending {
				s += " " + desc
			}

			query = query.Order(s)
		}
	}

	return query
}

func applyGetNamespaceInstancesFilters(query *gorm.DB, opts *instancestore.GetInstancesListOpts) *gorm.DB {
	for _, filter := range opts.Filter {
		switch filter.Field {
		case instancestore.FieldCreatedAt:
			if filter.Kind == instancestore.FilterKindBefore {
				query = query.Where(fieldCreatedAt+" < ?", filter.Value)
			} else if filter.Kind == instancestore.FilterKindAfter {
				query = query.Where(fieldCreatedAt+" > ?", filter.Value)
			} else {
				panic(fmt.Errorf("unexpected filter kind '%s' for use with field '%s'", filter.Kind, filter.Field))
			}

		case instancestore.FieldCalledAs:
			if filter.Kind == instancestore.FilterKindPrefix {
				query = query.Where(fieldCalledAs+" LIKE ?", filter.Value.(string)+"%")
			} else if filter.Kind == instancestore.FilterKindContains {
				query = query.Where(fieldCalledAs+" LIKE ?", "%"+filter.Value.(string)+"%")
			} else {
				panic(fmt.Errorf("unexpected filter kind '%s' for use with field '%s'", filter.Kind, filter.Field))
			}

		case instancestore.FieldStatus:
			if filter.Kind == instancestore.FilterKindMatch {
				query = query.Where(fieldStatus+" LIKE ?", filter.Value.(string))
			} else if filter.Kind == instancestore.FilterKindContains {
				query = query.Where(fieldStatus+" LIKE ?", "%"+filter.Value.(string)+"%")
			} else {
				panic(fmt.Errorf("unexpected filter kind '%s' for use with field '%s'", filter.Kind, filter.Field))
			}

		case instancestore.FieldInvoker:
			if filter.Kind == instancestore.FilterKindMatch {
				query = query.Where(fieldInvoker+" LIKE ?", filter.Value.(string))
			} else if filter.Kind == instancestore.FilterKindContains {
				query = query.Where(fieldInvoker+" LIKE ?", "%"+filter.Value.(string)+"%")
			} else {
				panic(fmt.Errorf("unexpected filter kind '%s' for use with field '%s'", filter.Kind, filter.Field))
			}

		default:
			panic(fmt.Errorf("unexpected filter field '%s'", filter.Field))
		}
	}

	return query
}

func (s *sqlInstanceStore) GetNamespaceInstances(ctx context.Context, nsID uuid.UUID, opts *instancestore.GetInstancesListOpts) (*instancestore.GetNamespaceInstancesResults, error) {
	var list []instancestore.InstanceData
	query := s.db.WithContext(ctx).Table(table).
		Select(summaryFields).
		Where(fieldNamespaceID, nsID)

	if opts != nil {
		if opts.LimitResults != 0 {
			query = query.Limit(opts.LimitResults)
		}

		if opts.OffsetResults != 0 {
			query = query.Offset(opts.OffsetResults)
		}

		query = applyGetNamespaceInstancesOrderings(query, opts)
		query = applyGetNamespaceInstancesFilters(query, opts)
	}

	res := query.Find(&list)
	if res.Error != nil {
		return nil, res.Error
	}

	var idatas []*instancestore.InstanceData
	for i := range list {
		idatas = append(idatas, &list[i])
	}

	return &instancestore.GetNamespaceInstancesResults{
		Total:   int(res.RowsAffected),
		Results: idatas,
	}, nil
}

func (s *sqlInstanceStore) GetHangingInstances(ctx context.Context) ([]*instancestore.InstanceData, error) {
	var list []instancestore.InstanceData
	query := s.db.WithContext(ctx).Table(table).
		Select(summaryFields).
		Where(fieldStatus+" < ?", instancestore.InstanceStatusComplete).
		Where(fieldDeadline+" < ?", time.Now())

	res := query.Find(&list)
	if res.Error != nil {
		return nil, res.Error
	}

	var idatas []*instancestore.InstanceData
	for i := range list {
		idatas = append(idatas, &list[i])
	}

	return idatas, nil
}

func (s *sqlInstanceStore) DeleteOldInstances(ctx context.Context, before time.Time) error {
	var idata instancestore.InstanceData
	res := s.db.WithContext(ctx).Table(table).
		Where(fieldStatus+" >= ?", instancestore.InstanceStatusComplete).
		Where(fieldEndedAt+" < ?", before).
		Delete(&idata)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (s *sqlInstanceStore) AssertNoParallelCron(ctx context.Context, wfID uuid.UUID) error {
	var idata instancestore.InstanceData
	res := s.db.WithContext(ctx).Table(table).
		Where(fieldInvoker+" = ?", instancestore.InvokerCron).
		Where(fieldWorkflowID+" = ?", wfID).
		Where(fieldCreatedAt+" > ?", time.Now().Add(time.Second*30)).
		First(&idata)
	if res.Error == nil && res.RowsAffected != 0 {
		return instancestore.ErrParallelCron
	}

	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil
	}

	return res.Error
}
