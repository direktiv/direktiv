package datastoresql

import (
	"context"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/logcollection"
	"gorm.io/gorm"
)

var _ logcollection.LogStore = &sqlLogNewStore{} // Ensures SQLLogStore struct conforms to logengine.LogStore interface.

type sqlLogNewStore struct {
	db *gorm.DB
}

func (s sqlLogNewStore) Get(ctx context.Context, stream string, offset int) ([]logcollection.LogEntry, error) {
	query := `
		SELECT time, tag, data
		FROM engine_messages
		WHERE stream = ?
		ORDER BY timestamp ASC LIMIT 200, OFFSET ?
	`
	resultList := make([]logcollection.LogEntry, 0)
	tx := s.db.WithContext(ctx).Raw(query).Scan(&resultList)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return resultList, nil
}

func (s sqlLogNewStore) GetInstanceLogs(ctx context.Context, stream string, instanceID string, offset int) ([]logcollection.LogEntry, error) {
	query := `
		SELECT time, tag, data
		FROM engine_messages
		WHERE stream = ? AND data->'entry'->>'callpath' LIKE ?
		ORDER BY timestamp ASC
		LIMIT 200 OFFSET ?
	`
	resultList := make([]logcollection.LogEntry, 0)
	tx := s.db.WithContext(ctx).Raw(query, stream, "%"+instanceID+"%", offset).Scan(&resultList)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return resultList, nil
}

func (s sqlLogNewStore) DeleteOldLogs(ctx context.Context, t time.Time) error {
	query := `
		DELETE FROM engine_messages
		WHERE time < ?
	`
	tx := s.db.WithContext(ctx).Exec(query, t)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}
