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
	table = "instances_v2"
)

var (
	summaryFields = []string{
		"id", "namespace_id", "workflow_id", "revision_id", "root_instance_id",
		"created_at", "updated_at", "ended_at", "deadline", "status", "called_as",
		"error_code", "invoker",
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

func (s *sqlInstanceStore) GetNamespaceInstances(ctx context.Context, nsID uuid.UUID) ([]*instancestore.InstanceData, error) {
	var list []instancestore.InstanceData
	res := s.db.WithContext(ctx).Table(table).
		Select(summaryFields).
		Where("namespace_id", nsID).
		Find(&list)
	if res.Error != nil {
		return nil, res.Error
	}

	var idatas []*instancestore.InstanceData
	for i := range list {
		idatas = append(idatas, &list[i])
	}

	return idatas, nil
}
