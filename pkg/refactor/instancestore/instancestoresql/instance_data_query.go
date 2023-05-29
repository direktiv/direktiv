package instancestoresql

import (
	"context"
	"errors"
	"fmt"

	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type instanceDataQuery struct {
	instanceID uuid.UUID
	db         *gorm.DB
}

var _ instancestore.InstanceDataQuery = &instanceDataQuery{} // Ensures instanceDataQuery struct conforms to instancestore.InstanceDataQuery interface.

func (q *instanceDataQuery) UpdateInstanceData(ctx context.Context, args *instancestore.UpdateInstanceDataArgs) (*instancestore.InstanceData, error) {
	idata := &instancestore.InstanceData{
		ID: q.instanceID,
	}

	m := make(map[string]interface{})

	if args.EndedAt != nil {
		m[fieldEndedAt] = *args.EndedAt
	}

	if args.Deadline != nil {
		m[fieldDeadline] = *args.Deadline
	}

	if args.Status != nil {
		m[fieldStatus] = *args.Status
	}

	if args.ErrorCode != nil {
		m[fieldErrorCode] = *args.ErrorCode
	}

	if args.TelemetryInfo != nil {
		m[fieldTelemetryInfo] = *args.TelemetryInfo
	}

	if args.RuntimeInfo != nil {
		m[fieldRuntimeInfo] = *args.RuntimeInfo
	}

	if args.ChildrenInfo != nil {
		m[fieldChildrenInfo] = *args.ChildrenInfo
	}

	if args.LiveData != nil {
		m[fieldLiveData] = *args.LiveData
	}

	if args.StateMemory != nil {
		m[fieldStateMemory] = *args.StateMemory
	}

	if args.Output != nil {
		m[fieldOutput] = *args.Output
	}

	if args.ErrorMessage != nil {
		m[fieldErrorMessage] = *args.ErrorMessage
	}

	if args.Metadata != nil {
		m[fieldMetadata] = *args.Metadata
	}

	res := q.db.WithContext(ctx).Model(&idata).Updates(m)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("instance '%s': %w", q.instanceID, instancestore.ErrNotFound)
		}

		return nil, res.Error
	}

	return idata, nil
}

func (q *instanceDataQuery) GetMost(ctx context.Context) (*instancestore.InstanceData, error) {
	idata := &instancestore.InstanceData{}

	res := q.db.WithContext(ctx).Table(table).
		Where("id", q.instanceID).
		Select(mostFields).
		First(idata)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("instance '%s': %w", q.instanceID, instancestore.ErrNotFound)
		}

		return nil, res.Error
	}

	return idata, nil
}

func (q *instanceDataQuery) GetSummary(ctx context.Context) (*instancestore.InstanceData, error) {
	idata := &instancestore.InstanceData{}

	res := q.db.WithContext(ctx).Table(table).
		Where("id", q.instanceID).
		Select(summaryFields).
		First(idata)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("instance '%s': %w", q.instanceID, instancestore.ErrNotFound)
		}

		return nil, res.Error
	}

	return idata, nil
}

func (q *instanceDataQuery) GetSummaryWithInput(ctx context.Context) (*instancestore.InstanceData, error) {
	idata := &instancestore.InstanceData{}

	res := q.db.WithContext(ctx).Table(table).
		Where("id", q.instanceID).
		Select(append(summaryFields, "input")).
		First(idata)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("instance '%s': %w", q.instanceID, instancestore.ErrNotFound)
		}

		return nil, res.Error
	}

	return idata, nil
}

func (q *instanceDataQuery) GetSummaryWithOutput(ctx context.Context) (*instancestore.InstanceData, error) {
	idata := &instancestore.InstanceData{}

	res := q.db.WithContext(ctx).Table(table).
		Where("id", q.instanceID).
		Select(append(summaryFields, "output")).
		First(idata)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("instance '%s': %w", q.instanceID, instancestore.ErrNotFound)
		}

		return nil, res.Error
	}

	return idata, nil
}

func (q *instanceDataQuery) GetSummaryWithMetadata(ctx context.Context) (*instancestore.InstanceData, error) {
	idata := &instancestore.InstanceData{}

	res := q.db.WithContext(ctx).Table(table).
		Where("id", q.instanceID).
		Select(append(summaryFields, "metadata")).
		First(idata)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("instance '%s': %w", q.instanceID, instancestore.ErrNotFound)
		}

		return nil, res.Error
	}

	return idata, nil
}
