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

func (s *sqlRuntimeVariablesStore) GetForNamespace(ctx context.Context, namespace string, name string) (*core.RuntimeVariable, error) {
	variable := &core.RuntimeVariable{}

	if name == "" || namespace == "" {
		return nil, datastore.ErrNotFound
	}

	res := s.db.WithContext(ctx).Raw(`
							SELECT 
								id, namespace, workflow_path, instance_id, 
								name, length(data) AS size, mime_type,
								created_at, updated_at
							FROM runtime_variables WHERE name = ? AND namespace = ? AND workflow_path IS NULL AND instance_id IS NULL`,
		name, namespace).First(variable)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, datastore.ErrNotFound
	}
	if res.Error != nil {
		return nil, res.Error
	}

	return variable, nil
}

func (s *sqlRuntimeVariablesStore) GetForInstance(ctx context.Context, instanceID uuid.UUID, name string) (*core.RuntimeVariable, error) {
	variable := &core.RuntimeVariable{}

	if name == "" || instanceID.String() == (uuid.UUID{}).String() {
		return nil, datastore.ErrNotFound
	}

	res := s.db.WithContext(ctx).Raw(`
							SELECT 
								id, namespace, workflow_path, instance_id, 
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

func (s *sqlRuntimeVariablesStore) GetForWorkflow(ctx context.Context, namespace string, path string, name string) (*core.RuntimeVariable, error) {
	variable := &core.RuntimeVariable{}

	if name == "" || path == "" || namespace == "" {
		return nil, datastore.ErrNotFound
	}

	res := s.db.WithContext(ctx).Raw(`
							SELECT 
								id, namespace, workflow_path, instance_id, 
								name, length(data) AS size, mime_type,
								created_at, updated_at
							FROM runtime_variables WHERE namespace = ? AND name = ? AND (workflow_path=?)`,
		namespace, name, path).First(variable)
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
								id, namespace, workflow_path, instance_id, 
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
	vals := make([]interface{}, 0)

	for idx, fieldName := range fieldNames {
		if fieldValues[idx] != nil {
			conditions = append(conditions, fmt.Sprintf(`"%s" = ?`, fieldName))
			vals = append(vals, fieldValues[idx])
		} else {
			conditions = append(conditions, fmt.Sprintf(`"%s" IS NULL`, fieldName))
		}
	}

	aggregateConditions := strings.Join(conditions, " AND ")

	res := s.db.WithContext(ctx).Raw(fmt.Sprintf(`
							SELECT 
								id, namespace, workflow_path, instance_id, 
								name, length(data) AS size, mime_type, 
								created_at, updated_at
							FROM runtime_variables WHERE %s`, aggregateConditions),
		vals...).Find(&variables)
	if res.Error != nil {
		return nil, res.Error
	}

	return variables, nil
}

func (s *sqlRuntimeVariablesStore) ListForInstance(ctx context.Context, instanceID uuid.UUID) ([]*core.RuntimeVariable, error) {
	return s.listByFieldValue(ctx, []string{"instance_id"}, []interface{}{instanceID.String()})
}

func (s *sqlRuntimeVariablesStore) ListForWorkflow(ctx context.Context, namespace string, workflowPath string) ([]*core.RuntimeVariable, error) {
	return s.listByFieldValue(ctx, []string{"namespace", "workflow_path"}, []interface{}{namespace, workflowPath})
}

func (s *sqlRuntimeVariablesStore) ListForNamespace(ctx context.Context, namespace string) ([]*core.RuntimeVariable, error) {
	return s.listByFieldValue(ctx, []string{"namespace", "workflow_path", "instance_id"}, []interface{}{namespace, nil, nil})
}

func (s *sqlRuntimeVariablesStore) get(ctx context.Context, variable *core.RuntimeVariable) (*core.RuntimeVariable, error) {
	if variable.WorkflowPath != "" {
		return s.GetForWorkflow(ctx, variable.Namespace, variable.WorkflowPath, variable.Name)
	}

	zeroUUID := (uuid.UUID{}).String()

	if variable.InstanceID.String() != zeroUUID {
		return s.GetForInstance(ctx, variable.InstanceID, variable.Name)
	}

	return s.GetForNamespace(ctx, variable.Namespace, variable.Name)
}

func (s *sqlRuntimeVariablesStore) Set(ctx context.Context, variable *core.RuntimeVariable) (*core.RuntimeVariable, error) {
	if variable.Name == "" {
		return nil, core.ErrInvalidRuntimeVariableName
	}
	if matched, _ := regexp.MatchString(core.RuntimeVariableNameRegexPattern, variable.Name); !matched {
		return nil, core.ErrInvalidRuntimeVariableName
	}
	zeroUUID := (uuid.UUID{}).String()
	if variable.InstanceID.String() == zeroUUID && variable.Namespace == "" && variable.WorkflowPath == "" {
		return nil, core.ErrInvalidRuntimeVariableName
	}

	selectorField := ""

	var extra string
	args := []interface{}{
		variable.MimeType, variable.Data, variable.Namespace, variable.Name,
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
		extra = "AND workflow_path IS NULL AND instance_id IS NULL"
		// args = append(args, nil, nil)
	}

	queryString := fmt.Sprintf(
		`UPDATE runtime_variables SET
						mime_type=?,
						data=?
					WHERE namespace = ? AND name = ? %s;`, extra)

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
	}

	newUUID := uuid.New()
	args = append([]interface{}{newUUID}, args...)

	res = s.db.WithContext(ctx).Exec(fmt.Sprintf(`
							INSERT INTO runtime_variables(
								id, mime_type, data, namespace, name%s) 
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
		`UPDATE runtime_variables SET name=? WHERE id=?`,
		name, id)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, datastore.ErrNotFound
	}
	if res.RowsAffected > 1 {
		return nil, fmt.Errorf("unexpected runtime_variables update count, got: %d, want: %d", res.RowsAffected, 1)
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
		return fmt.Errorf("unexpected runtime_variables delete count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}

func (s *sqlRuntimeVariablesStore) LoadData(ctx context.Context, id uuid.UUID) ([]byte, error) {
	variable := &core.RuntimeVariable{}
	res := s.db.WithContext(ctx).Raw(`
							SELECT 
								id, namespace, workflow_path, instance_id, 
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

func (s *sqlRuntimeVariablesStore) DeleteForWorkflow(ctx context.Context, namespace string, workflowPath string) error {
	res := s.db.WithContext(ctx).Exec(
		`DELETE FROM runtime_variables WHERE namespace=? AND workflow_path=?`,
		namespace, workflowPath)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (s *sqlRuntimeVariablesStore) SetWorkflowPath(ctx context.Context, namespace string, oldWorkflowPath string, newWorkflowPath string) error {
	res := s.db.WithContext(ctx).Exec(
		`UPDATE runtime_variables SET workflow_path=? WHERE namespace=? AND workflow_path=?`,
		newWorkflowPath, namespace, oldWorkflowPath)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
