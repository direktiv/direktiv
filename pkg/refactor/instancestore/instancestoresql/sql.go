package instancestoresql

import (
	"context"
	"fmt"

	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	table          = "instances_v2"
	fieldCreatedAt = "created_at"
	fieldCalledAs  = "called_as"
	fieldStatus    = "status"
	fieldInvoker   = "invoker"
	desc           = "desc"
)

var (
	summaryFields = []string{
		"id", "namespace_id", "workflow_id", "revision_id", "root_instance_id",
		fieldCreatedAt, "updated_at", "ended_at", "deadline", fieldStatus, fieldCalledAs,
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
		Status:         util.InstanceStatusPending,
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

func (s *sqlInstanceStore) GetNamespaceInstances(ctx context.Context, nsID uuid.UUID, opts *instancestore.GetInstancesListOpts) ([]*instancestore.InstanceData, error) {
	var list []instancestore.InstanceData
	query := s.db.WithContext(ctx).Table(table).
		Select(summaryFields).
		Where("namespace_id", nsID)

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

	return idatas, nil
}
