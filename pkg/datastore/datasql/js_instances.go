package datasql

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sqlJSInstancesStore struct {
	db *gorm.DB
}

func (s *sqlJSInstancesStore) Patch(ctx context.Context, id uuid.UUID, patch map[string]any) error {
	query := "UPDATE js_instances SET updated_at=NOW()"
	setValues := []any{}

	allowedFields := ",status,memory,output,error,ended_at,"

	for key, value := range patch {
		if !strings.Contains(allowedFields, ","+key+",") {
			return datastore.ErrInvalidArgument
		}
		if key == "ended_at" {
			query += ", ended_at=NOW()"
			continue
		}
		query += ", " + key + "=?"
		setValues = append(setValues, value)
	}
	if len(setValues) < 1 {
		return datastore.ErrInvalidArgument
	}
	query += " WHERE id=?"
	setValues = append(setValues, id)

	res := s.db.WithContext(ctx).Exec(query,
		setValues...)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpected gorm update count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}

func (s *sqlJSInstancesStore) GetByID(ctx context.Context, id uuid.UUID) (*datastore.JSInstance, error) {
	obj := &datastore.JSInstance{}
	res := s.db.WithContext(ctx).Raw(`
							SELECT	id, namespace, workflow_path, workflow_data,
									status,
									input, memory, output, error,
									created_at, updated_at, ended_at

							FROM js_instances
							WHERE id=?`,
		id).
		First(obj)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, datastore.ErrNotFound
	}
	if res.Error != nil {
		return nil, res.Error
	}

	return obj, nil
}

func (s *sqlJSInstancesStore) Create(ctx context.Context, jsInstance *datastore.JSInstance) error {
	res := s.db.WithContext(ctx).Exec(`
				INSERT INTO js_instances(
									id, namespace, workflow_path, workflow_data, 
									status, 
									input, memory, output, error,
									created_at, updated_at) 
									VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW());`,
		jsInstance.ID, jsInstance.Namespace, jsInstance.WorkflowPath, jsInstance.WorkflowData,
		jsInstance.Status,
		jsInstance.Input, jsInstance.Memory, jsInstance.Output, jsInstance.Error)

	if res.Error != nil {
		return res.Error
	}

	return nil
}

var _ datastore.JSInstancesStore = &sqlJSInstancesStore{}
