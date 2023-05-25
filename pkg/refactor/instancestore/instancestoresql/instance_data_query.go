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
	// TODO: alan
	return nil, nil
}

func (q *instanceDataQuery) GetEverything(ctx context.Context) (*instancestore.InstanceData, error) {
	idata := &instancestore.InstanceData{}

	res := q.db.WithContext(ctx).Table(table).
		Where("id", q.instanceID).
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
