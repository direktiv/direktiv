package instancestoresql

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/instancestore"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	maxWriteDT = time.Minute / 2
)

type instanceDataQuery struct {
	instanceID uuid.UUID
	db         *gorm.DB
}

var _ instancestore.InstanceDataQuery = &instanceDataQuery{} // Ensures instanceDataQuery struct conforms to instancestore.InstanceDataQuery interface.

func (q *instanceDataQuery) UpdateInstanceData(ctx context.Context, args *instancestore.UpdateInstanceDataArgs) error {
	var vals []interface{}
	var clauses []string
	query := fmt.Sprintf("UPDATE %s", table)

	clauses = append(clauses, fmt.Sprintf("%s = ?", fieldUpdatedAt))
	vals = append(vals, time.Now().UTC())

	clauses = append(clauses, fmt.Sprintf("%s = ?", fieldServer))
	vals = append(vals, args.Server.String())

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

	query += fmt.Sprintf(" SET %s", strings.Join(clauses, ", "))

	query += fmt.Sprintf(" WHERE %s = ?", fieldID)
	vals = append(vals, q.instanceID)

	if !args.BypassOwnershipCheck {
		query += fmt.Sprintf(" AND %s = ? AND %s > ?", fieldServer, fieldUpdatedAt)
		vals = append(vals, args.Server)
		vals = append(vals, time.Now().UTC().Add(-maxWriteDT))
	}

	res := q.db.WithContext(ctx).Exec(query, vals...)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("instance '%s' not found or not allowed to write: %w", q.instanceID, instancestore.ErrNotFound)
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

func (q *instanceDataQuery) EnqueueMessage(ctx context.Context, args *instancestore.EnqueueInstanceMessageArgs) error {
	idata := &instancestore.InstanceMessageData{
		ID:         uuid.New(),
		InstanceID: q.instanceID,
		Payload:    args.Payload,
	}

	columns := []string{
		fieldInstanceMessageID, fieldInstanceMessageInstanceID, fieldInstanceMessagePayload,
	}
	query := generateInsertQuery(messagesTable, columns)

	res := q.db.WithContext(ctx).Exec(query,
		idata.ID, idata.InstanceID, idata.Payload)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpected gorm create count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}

func (q *instanceDataQuery) PopMessage(ctx context.Context) (*instancestore.InstanceMessageData, error) {
	columns := []string{fieldInstanceMessageID, fieldInstanceMessageInstanceID, fieldInstanceMessageCreatedAt, fieldInstanceMessagePayload}
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE %s = ? ORDER BY %s ASC LIMIT 1`, strings.Join(columns, ", "), messagesTable, fieldInstanceMessageInstanceID, fieldInstanceMessageCreatedAt)

	msg := &instancestore.InstanceMessageData{}
	res := q.db.WithContext(ctx).Raw(query, q.instanceID).First(msg)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, instancestore.ErrNoMessages
		}

		return nil, res.Error
	}

	res = q.db.WithContext(ctx).Exec(fmt.Sprintf(`DELETE FROM %s WHERE %s = ?`, messagesTable, fieldInstanceMessageID), msg.ID)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, instancestore.ErrNoMessages
	}
	if res.RowsAffected > 1 {
		return nil, fmt.Errorf("unexpected instance_messages delete count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return msg, nil
}
