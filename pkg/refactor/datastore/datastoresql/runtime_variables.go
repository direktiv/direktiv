package datastoresql

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sqlRuntimeVariablesStore struct {
	db *gorm.DB
}

func (s *sqlRuntimeVariablesStore) GetByReferenceAndName(ctx context.Context, referenceID uuid.UUID, name string) (*core.RuntimeVariable, error) {
	variable := &core.RuntimeVariable{}
	res := s.db.WithContext(ctx).Raw(`
							SELECT 
								id, namespace_id, workflow_id, instance_id, 
								name, length(data) AS size, mime_type,
								created_at, updated_at
							FROM runtime_variables WHERE name = ? AND (namespace_id=? OR workflow_id=? OR instance_id=?);`,
		name, referenceID, referenceID, referenceID).First(variable)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, datastore.ErrNotFound
	}
	if res.Error != nil {
		return nil, res.Error
	}

	return variable, nil
}

func (s *sqlRuntimeVariablesStore) GetByID(ctx context.Context, id uuid.UUID) (*core.RuntimeVariable, error) {
	variable := &core.RuntimeVariable{}
	res := s.db.WithContext(ctx).Raw(`
							SELECT 
								id, namespace_id, workflow_id, instance_id, 
								name, length(data) AS size, mime_type,
								created_at, updated_at
							FROM runtime_variables WHERE "id" = ?;`,
		id).First(variable)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, datastore.ErrNotFound
	}
	if res.Error != nil {
		return nil, res.Error
	}

	return variable, nil
}

func (s *sqlRuntimeVariablesStore) listByFieldID(ctx context.Context, fieldName string, fieldID uuid.UUID) ([]*core.RuntimeVariable, error) {
	var variables []*core.RuntimeVariable

	res := s.db.WithContext(ctx).Raw(fmt.Sprintf(`
							SELECT 
								id, namespace_id, workflow_id, instance_id, 
								name, length(data) AS size, mime_type, 
								created_at, updated_at
							FROM runtime_variables WHERE "%s" = ?`, fieldName),
		fieldID).Find(&variables)
	if res.Error != nil {
		return nil, res.Error
	}

	return variables, nil
}

func (s *sqlRuntimeVariablesStore) ListByInstanceID(ctx context.Context, instanceID uuid.UUID) ([]*core.RuntimeVariable, error) {
	return s.listByFieldID(ctx, "instance_id", instanceID)
}

func (s *sqlRuntimeVariablesStore) ListByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]*core.RuntimeVariable, error) {
	return s.listByFieldID(ctx, "workflow_id", workflowID)
}

func (s *sqlRuntimeVariablesStore) ListByNamespaceID(ctx context.Context, namespaceID uuid.UUID) ([]*core.RuntimeVariable, error) {
	return s.listByFieldID(ctx, "namespace_id", namespaceID)
}

func (s *sqlRuntimeVariablesStore) Set(ctx context.Context, variable *core.RuntimeVariable) (*core.RuntimeVariable, error) {
	if matched, _ := regexp.MatchString(core.RuntimeVariableNameRegexPattern, variable.Name); !matched {
		return nil, core.ErrInvalidRuntimeVariableName
	}
	linkName := "namespace_id"
	linkValue := variable.NamespaceID

	if variable.WorkflowID.String() != (uuid.UUID{}).String() {
		linkName = "workflow_id"
		linkValue = variable.WorkflowID
	}

	if variable.InstanceID.String() != (uuid.UUID{}).String() {
		linkName = "instance_id"
		linkValue = variable.InstanceID
	}

	res := s.db.WithContext(ctx).Exec(fmt.Sprintf(
		`UPDATE runtime_variables SET
						mime_type=?,
						data=?
					WHERE %s = ? AND name = ?;`, linkName),
		variable.MimeType, variable.Data, linkValue, variable.Name)

	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected > 1 {
		return nil, fmt.Errorf("unexpected gorm update count, got: %d, want: %d", res.RowsAffected, 1)
	}
	if res.RowsAffected == 1 {
		return s.GetByReferenceAndName(ctx, linkValue, variable.Name)
	}

	newUUID := uuid.New()
	res = s.db.WithContext(ctx).Exec(fmt.Sprintf(`
							INSERT INTO runtime_variables(
								id, %s, name, mime_type, data) 
							VALUES(?, ?, ?, ?, ?);`, linkName),
		newUUID, linkValue, variable.Name, variable.MimeType, variable.Data)

	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpected gorm insert count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return s.GetByID(ctx, newUUID)
}

func (s *sqlRuntimeVariablesStore) SetName(ctx context.Context, id uuid.UUID, name string) (*core.RuntimeVariable, error) {
	if matched, _ := regexp.MatchString(core.RuntimeVariableNameRegexPattern, name); !matched {
		return nil, core.ErrInvalidRuntimeVariableName
	}
	res := s.db.WithContext(ctx).Exec(
		`UPDATE runtime_variables SET name=? WHERE id = ?`,
		name, id)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, datastore.ErrNotFound
	}
	if res.RowsAffected > 1 {
		return nil, fmt.Errorf("unexpedted runtime_variables update count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return s.GetByID(ctx, id)
}

func (s *sqlRuntimeVariablesStore) Delete(ctx context.Context, id uuid.UUID) error {
	res := s.db.WithContext(ctx).Exec(
		`DELETE FROM runtime_variables WHERE id = ?;`,
		id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return datastore.ErrNotFound
	}
	if res.RowsAffected > 1 {
		return fmt.Errorf("unexpedted runtime_variables delete count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}

func (s *sqlRuntimeVariablesStore) LoadData(ctx context.Context, id uuid.UUID) ([]byte, error) {
	variable := &core.RuntimeVariable{}
	res := s.db.WithContext(ctx).Raw(`
							SELECT 
								id, namespace_id, workflow_id, instance_id, 
								name, length(data) AS size, mime_type, data,
								created_at, updated_at
							FROM runtime_variables WHERE "id" = ?;`,
		id).First(variable)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, datastore.ErrNotFound
	}
	if res.Error != nil {
		return nil, res.Error
	}

	return variable.Data, nil
}
