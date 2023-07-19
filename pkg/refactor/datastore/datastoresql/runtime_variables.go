package datastoresql

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sqlRuntimeVariablesStore struct {
	db *gorm.DB
}

func (s *sqlRuntimeVariablesStore) GetByNamespaceAndName(ctx context.Context, namespaceID uuid.UUID, name string) (*core.RuntimeVariable, error) {
	variable := &core.RuntimeVariable{}

	if name == "" || namespaceID.String() == (uuid.UUID{}).String() {
		return nil, datastore.ErrNotFound
	}

	res := s.db.WithContext(ctx).Raw(`
							SELECT 
								id, namespace_id, workflow_path, instance_id, 
								name, length(data) AS size, mime_type,
								created_at, updated_at
							FROM runtime_variables WHERE name = ? AND namespace_id = ? AND workflow_path IS NULL AND instance_id IS NULL`,
		name, namespaceID.String()).First(variable)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, datastore.ErrNotFound
	}
	if res.Error != nil {
		return nil, res.Error
	}

	return variable, nil
}

func (s *sqlRuntimeVariablesStore) GetByInstanceAndName(ctx context.Context, instanceID uuid.UUID, name string) (*core.RuntimeVariable, error) {
	variable := &core.RuntimeVariable{}

	if name == "" || instanceID.String() == (uuid.UUID{}).String() {
		return nil, datastore.ErrNotFound
	}

	res := s.db.WithContext(ctx).Raw(`
							SELECT 
								id, namespace_id, workflow_path, instance_id, 
								name, length(data) AS size, mime_type,
								created_at, updated_at
							FROM runtime_variables WHERE name = ? AND (instance_id=?)`,
		name, instanceID.String()).First(variable)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, datastore.ErrNotFound
	}
	if res.Error != nil {
		return nil, res.Error
	}

	return variable, nil
}

func (s *sqlRuntimeVariablesStore) GetByWorkflowAndName(ctx context.Context, namespaceID uuid.UUID, path string, name string) (*core.RuntimeVariable, error) {
	variable := &core.RuntimeVariable{}

	zeroUUID := (uuid.UUID{}).String()

	if name == "" || path == "" || namespaceID.String() == zeroUUID {
		return nil, datastore.ErrNotFound
	}

	res := s.db.WithContext(ctx).Raw(`
							SELECT 
								id, namespace_id, workflow_path, instance_id, 
								name, length(data) AS size, mime_type,
								created_at, updated_at
							FROM runtime_variables WHERE namespace_id = ? AND name = ? AND (workflow_path=?)`,
		namespaceID.String(), name, path).First(variable)
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
								id, namespace_id, workflow_path, instance_id, 
								name, length(data) AS size, mime_type,
								created_at, updated_at
							FROM runtime_variables WHERE "id" = ?`,
		id).First(variable)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, datastore.ErrNotFound
	}
	if res.Error != nil {
		return nil, res.Error
	}

	return variable, nil
}

func (s *sqlRuntimeVariablesStore) listByFieldValue(ctx context.Context, fieldNames []string, fieldValues []interface{}) ([]*core.RuntimeVariable, error) {
	var variables []*core.RuntimeVariable

	conditions := make([]string, 0)
	for _, fieldName := range fieldNames {
		conditions = append(conditions, fmt.Sprintf(`"%s" = ?`, fieldName))
	}

	aggregateConditions := strings.Join(conditions, " AND ")

	res := s.db.WithContext(ctx).Raw(fmt.Sprintf(`
							SELECT 
								id, namespace_id, workflow_path, instance_id, 
								name, length(data) AS size, mime_type, 
								created_at, updated_at
							FROM runtime_variables WHERE %s`, aggregateConditions),
		fieldValues...).Find(&variables)
	if res.Error != nil {
		return nil, res.Error
	}

	return variables, nil
}

func (s *sqlRuntimeVariablesStore) ListByInstanceID(ctx context.Context, instanceID uuid.UUID) ([]*core.RuntimeVariable, error) {
	return s.listByFieldValue(ctx, []string{"instance_id"}, []interface{}{instanceID.String()})
}

func (s *sqlRuntimeVariablesStore) ListByWorkflowPath(ctx context.Context, namespaceID uuid.UUID, workflowPath string) ([]*core.RuntimeVariable, error) {
	return s.listByFieldValue(ctx, []string{"namespace_id", "workflow_path"}, []interface{}{namespaceID.String(), workflowPath})
}

func (s *sqlRuntimeVariablesStore) ListByNamespaceID(ctx context.Context, namespaceID uuid.UUID) ([]*core.RuntimeVariable, error) {
	return s.listByFieldValue(ctx, []string{"namespace_id"}, []interface{}{namespaceID.String()})
}

func (s *sqlRuntimeVariablesStore) get(ctx context.Context, variable *core.RuntimeVariable) (*core.RuntimeVariable, error) {
	if variable.WorkflowPath != "" {
		return s.GetByWorkflowAndName(ctx, variable.NamespaceID, variable.WorkflowPath, variable.Name)
	}

	zeroUUID := (uuid.UUID{}).String()

	if variable.InstanceID.String() != zeroUUID {
		return s.GetByInstanceAndName(ctx, variable.InstanceID, variable.Name)
	}

	return s.GetByNamespaceAndName(ctx, variable.NamespaceID, variable.Name)
}

func (s *sqlRuntimeVariablesStore) Set(ctx context.Context, variable *core.RuntimeVariable) (*core.RuntimeVariable, error) {
	if variable.Name == "" {
		return nil, core.ErrInvalidRuntimeVariableName
	}
	if matched, _ := regexp.MatchString(core.RuntimeVariableNameRegexPattern, variable.Name); !matched {
		return nil, core.ErrInvalidRuntimeVariableName
	}
	zeroUUID := (uuid.UUID{}).String()
	if variable.InstanceID.String() == zeroUUID && variable.NamespaceID.String() == zeroUUID && variable.WorkflowPath == "" {
		return nil, core.ErrInvalidRuntimeVariableName
	}

	selectorField := ""

	extra := ""
	args := []interface{}{
		variable.MimeType, variable.Data, variable.NamespaceID.String(), variable.Name,
	}

	if variable.InstanceID.String() != zeroUUID {
		selectorField = "instance_id"
		extra = fmt.Sprintf("AND %s = ?", selectorField)
		args = append(args, variable.InstanceID.String())
	} else if variable.WorkflowPath != "" {
		selectorField = "workflow_path"
		extra = fmt.Sprintf("AND %s = ?", selectorField)
		args = append(args, variable.WorkflowPath)
	} else {
		extra = fmt.Sprintf("AND workflow_path IS NULL AND instance_id IS NULL")
		// args = append(args, nil, nil)
	}

	queryString := fmt.Sprintf(
		`UPDATE runtime_variables SET
						mime_type=?,
						data=?
					WHERE namespace_id = ? AND name = ? %s;`, extra)

	res := s.db.WithContext(ctx).Exec(queryString, args...)

	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected > 1 {
		return nil, fmt.Errorf("unexpected gorm update count, got: %d, want: %d", res.RowsAffected, 1)
	}
	if res.RowsAffected == 1 {
		return s.get(ctx, variable)
	}

	extraVal := ""

	if selectorField != "" {
		selectorField = ", " + selectorField
		extraVal = ", ?"
	} else {
		// args = args[:len(args)-2]
		// selectorField = fmt.Sprintf(`, workflow_path, instance_id`)
		// extraVal = ", ?, ?"
	}

	newUUID := uuid.New()
	args = append([]interface{}{newUUID}, args...)

	res = s.db.WithContext(ctx).Exec(fmt.Sprintf(`
							INSERT INTO runtime_variables(
								id, mime_type, data, namespace_id, name%s) 
							VALUES(?, ?, ?, ?, ?%s);`, selectorField, extraVal),
		args...)

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
		`DELETE FROM runtime_variables WHERE id = ?`,
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
								id, namespace_id, workflow_path, instance_id, 
								name, length(data) AS size, mime_type, data,
								created_at, updated_at
							FROM runtime_variables WHERE "id" = ?`,
		id).First(variable)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, datastore.ErrNotFound
	}
	if res.Error != nil {
		return nil, res.Error
	}

	return variable.Data, nil
}
