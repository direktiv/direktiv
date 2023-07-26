package instancestoresql

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type instanceDataQuery struct {
	instanceID uuid.UUID
	db         *gorm.DB
	logger     *zap.SugaredLogger
}

var _ instancestore.InstanceDataQuery = &instanceDataQuery{} // Ensures instanceDataQuery struct conforms to instancestore.InstanceDataQuery interface.

func (q *instanceDataQuery) UpdateInstanceData(ctx context.Context, args *instancestore.UpdateInstanceDataArgs) error {
	var vals []interface{}
	var clauses []string
	query := fmt.Sprintf("UPDATE %s", table)

	if args.EndedAt != nil {
		clauses = append(clauses, fmt.Sprintf("%s = ?", fieldEndedAt))
		vals = append(vals, args.EndedAt.UTC())
	}

	if args.Deadline != nil {
		clauses = append(clauses, fmt.Sprintf("%s = ?", fieldDeadline))
		vals = append(vals, args.Deadline.UTC())
	}

	if args.Status != nil {
		clauses = append(clauses, fmt.Sprintf("%s = ?", fieldStatus))
		vals = append(vals, *args.Status)
	}

	if args.ErrorCode != nil {
		clauses = append(clauses, fmt.Sprintf("%s = ?", fieldErrorCode))
		vals = append(vals, *args.ErrorCode)
	}

	if args.TelemetryInfo != nil {
		clauses = append(clauses, fmt.Sprintf("%s = ?", fieldTelemetryInfo))
		vals = append(vals, *args.TelemetryInfo)
	}

	if args.RuntimeInfo != nil {
		clauses = append(clauses, fmt.Sprintf("%s = ?", fieldRuntimeInfo))
		vals = append(vals, *args.RuntimeInfo)
	}

	if args.ChildrenInfo != nil {
		clauses = append(clauses, fmt.Sprintf("%s = ?", fieldChildrenInfo))
		vals = append(vals, *args.ChildrenInfo)
	}

	if args.LiveData != nil {
		clauses = append(clauses, fmt.Sprintf("%s = ?", fieldLiveData))
		vals = append(vals, *args.LiveData)
	}

	if args.StateMemory != nil {
		clauses = append(clauses, fmt.Sprintf("%s = ?", fieldStateMemory))
		vals = append(vals, *args.StateMemory)
	}

	if args.Output != nil {
		clauses = append(clauses, fmt.Sprintf("%s = ?", fieldOutput))
		vals = append(vals, *args.Output)
	}

	if args.ErrorMessage != nil {
		clauses = append(clauses, fmt.Sprintf("%s = ?", fieldErrorMessage))
		vals = append(vals, *args.ErrorMessage)
	}

	if args.Metadata != nil {
		clauses = append(clauses, fmt.Sprintf("%s = ?", fieldMetadata))
		vals = append(vals, *args.Metadata)
	}

	// if len(clauses) == 0 {
	// 	return
	// }

	query += fmt.Sprintf(" SET %s", strings.Join(clauses, ", "))

	query += fmt.Sprintf(" WHERE %s = ?", fieldID)

	q.logger.Debug(fmt.Sprintf("UpdateInstanceData executing SQL query: %s", query))

	vals = append(vals, q.instanceID)

	res := q.db.WithContext(ctx).Exec(query, vals...)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("instance '%s': %w", q.instanceID, instancestore.ErrNotFound)
		}

		return res.Error
	}

	if res.RowsAffected != 1 {
		return fmt.Errorf("UpdateInstanceData expected 1 rows affected, got %d", res.RowsAffected)
	}

	return nil
}

func (q *instanceDataQuery) get(ctx context.Context, columns []string) (*instancestore.InstanceData, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE %s = ?`, strings.Join(columns, ", "), table, fieldID)

	q.logger.Debug(fmt.Sprintf("get executing SQL query: %s", query))

	idata := &instancestore.InstanceData{}
	res := q.db.WithContext(ctx).Raw(query, q.instanceID).First(idata)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("instance '%s': %w", q.instanceID, instancestore.ErrNotFound)
		}

		return nil, res.Error
	}

	return idata, nil
}

func (q *instanceDataQuery) GetMost(ctx context.Context) (*instancestore.InstanceData, error) {
	return q.get(ctx, mostFields)
}

func (q *instanceDataQuery) GetSummary(ctx context.Context) (*instancestore.InstanceData, error) {
	return q.get(ctx, summaryFields)
}

func (q *instanceDataQuery) GetSummaryWithInput(ctx context.Context) (*instancestore.InstanceData, error) {
	return q.get(ctx, append(summaryFields, fieldInput))
}

func (q *instanceDataQuery) GetSummaryWithOutput(ctx context.Context) (*instancestore.InstanceData, error) {
	return q.get(ctx, append(summaryFields, fieldOutput))
}

func (q *instanceDataQuery) GetSummaryWithMetadata(ctx context.Context) (*instancestore.InstanceData, error) {
	return q.get(ctx, append(summaryFields, fieldMetadata))
}
