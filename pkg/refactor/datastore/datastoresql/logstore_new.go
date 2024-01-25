package datastoresql

import (
	"context"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/logcollection"
	"gorm.io/gorm"
)

var _ logcollection.LogStore = &sqlLogNewStore{}
const pageSize = 200

type sqlLogNewStore struct {
	db *gorm.DB
}


func (s sqlLogNewStore) Get(ctx context.Context, stream string, cursorTime time.Time) ([]logcollection.LogEntry, error) {
	query := `
		SELECT time, tag, data
		FROM fluentbit
		WHERE tag = ? AND time > ?
		ORDER BY time ASC
		LIMIT ?;
	`
	resultList := make([]logcollection.LogEntry, 0)
	tx := s.db.WithContext(ctx).Raw(query, stream, cursorTime, pageSize).Scan(&resultList)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return resultList, nil
}

func (s sqlLogNewStore) GetInstanceLogs(ctx context.Context, stream string, instanceID string, cursorTime time.Time) ([]logcollection.LogEntry, error) {
	query := `
		SELECT time, tag, data
		FROM fluentbit
		WHERE tag = ? AND data->'entry'->>'callpath' LIKE ? AND time > ?
		ORDER BY time ASC
		LIMIT ?;
	`
	resultList := make([]logcollection.LogEntry, 0)
	tx := s.db.WithContext(ctx).Raw(query, stream, "%"+instanceID+"%", cursorTime, pageSize).Scan(&resultList)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return resultList, nil
}

func (s sqlLogNewStore) DeleteOldLogs(ctx context.Context, t time.Time) error {
	query := `
		DELETE FROM fluentbit
		WHERE time < ?
	`
	tx := s.db.WithContext(ctx).Exec(query, t)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}
