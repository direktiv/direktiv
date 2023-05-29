package datastoresql

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sqlRuntimeVariablesStore struct {
	db *gorm.DB
}

func (s *sqlRuntimeVariablesStore) GetByID(ctx context.Context, id uuid.UUID) (*core.RuntimeVariable, error) {
	variable := &core.RuntimeVariable{}
	res := s.db.WithContext(ctx).Raw(`
							SELECT 
								id, namespace_id, workflow_id, instance_id, 
								scope, name, size, hash, mime_type,
								created_at, updated_at
							FROM runtime_variables WHERE "id" = ?;`,
		id).First(variable)
	if res.Error != nil {
		return nil, res.Error
	}

	return variable, nil
}

func (s *sqlRuntimeVariablesStore) listByFieldID(ctx context.Context, fieldName string, fieldID uuid.UUID) (core.RuntimeVariablesList, error) {
	var variables []*core.RuntimeVariable

	res := s.db.WithContext(ctx).Raw(fmt.Sprintf(`
							SELECT 
								id, namespace_id, workflow_id, instance_id, 
								scope, name, size, hash, mime_type, 
								created_at, updated_at
							FROM runtime_variables WHERE "%s" = ?`, fieldName),
		fieldID).Find(&variables)
	if res.Error != nil {
		return nil, res.Error
	}

	return variables, nil
}

func (s *sqlRuntimeVariablesStore) ListByInstanceID(ctx context.Context, instanceID uuid.UUID) (core.RuntimeVariablesList, error) {
	return s.listByFieldID(ctx, "instance_id", instanceID)
}

func (s *sqlRuntimeVariablesStore) ListByWorkflowID(ctx context.Context, workflowID uuid.UUID) (core.RuntimeVariablesList, error) {
	return s.listByFieldID(ctx, "workflow_id", workflowID)
}

func (s *sqlRuntimeVariablesStore) ListByNamespaceID(ctx context.Context, namespaceID uuid.UUID) (core.RuntimeVariablesList, error) {
	return s.listByFieldID(ctx, "namespace_id", namespaceID)
}

func (s *sqlRuntimeVariablesStore) Set(ctx context.Context, variable *core.RuntimeVariable) (*core.RuntimeVariable, error) {
	hash := sha256.Sum256(variable.Data)
	size := len(variable.Data)

	linkName := "namespace_id"
	linkValue := variable.NamespaceID

	if variable.WorkflowID.String() != uuid.New().String() {
		linkName = "workflow_id"
		linkValue = variable.WorkflowID
	}
	if variable.InstanceID.String() != uuid.New().String() {
		linkName = "instance"
		linkValue = variable.InstanceID
	}

	res := s.db.WithContext(ctx).Exec(fmt.Sprintf(
		`UPDATE runtime_variables SET
						%s=?,
						scope=?, 
						name=?, 
						size=?, 
						hash=?, 
						mime_type=?
					WHERE id = ?;`, linkName),
		linkValue, variable.Scope, variable.Name, size, hash, variable.MimeType, variable.ID)

	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected > 1 {
		return nil, fmt.Errorf("unexpected gorm update count, got: %d, want: %d", res.RowsAffected, 1)
	}
	if res.RowsAffected == 1 {
		return s.GetByID(ctx, variable.ID)
	}

	res = s.db.WithContext(ctx).Exec(fmt.Sprintf(`
							INSERT INTO runtime_variables(
								id, %s, scope, name, size, hash, mime_type) 
							VALUES(?, ?, ?, ?, ?, ?, ?);`, linkName),
		uuid.New(), linkValue, variable.Scope, variable.Name, size, hash, variable.MimeType)

	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpected gorm insert count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return s.GetByID(ctx, variable.ID)
}

func (s *sqlRuntimeVariablesStore) SetName(ctx context.Context, id uuid.UUID, name string) (*core.RuntimeVariable, error) {
	res := s.db.WithContext(ctx).Exec(
		`UPDATE runtime_variables SET name=? WHERE id = ?`,
		name, id)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpedted gorm delete count, got: %d, want: %d", res.RowsAffected, 1)
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
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpedted gorm delete count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}

func (s *sqlRuntimeVariablesStore) LoadData(ctx context.Context, id uuid.UUID) ([]byte, error) {
	variable := &core.RuntimeVariable{}
	res := s.db.WithContext(ctx).Raw(`
							SELECT 
								id, namespace_id, workflow_id, instance_id, 
								scope, name, size, hash, mime_type, data,
								created_at, updated_at
							FROM runtime_variables WHERE "id" = ?;`,
		id).First(variable)
	if res.Error != nil {
		return nil, res.Error
	}

	return variable.Data, nil
}
