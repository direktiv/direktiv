package datastoresql

import (
	"context"
	"encoding/json"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/plattformlogs"
	"gorm.io/gorm"
)

var _ plattformlogs.LogStore = &sqlLogNewStore{}

const pageSize = 200

type sqlLogNewStore struct {
	db *gorm.DB
}

type ScanResult struct {
	ID   int
	Time time.Time
	Tag  string
	Data []byte
}

func (s sqlLogNewStore) GetOlder(ctx context.Context, track string, t time.Time) ([]plattformlogs.LogEntry, error) {
	query := `
        SELECT id, time, tag, data
        FROM fluentbit
        WHERE tag = ? AND time < ?
        ORDER BY time DESC
        LIMIT ?;
    `
	resultList := make([]ScanResult, 0)
	tx := s.db.WithContext(ctx).Raw(query, track, t, pageSize).Scan(&resultList)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return convertScanResults(resultList)
}

func (s sqlLogNewStore) GetOlderInstance(ctx context.Context, track string, t time.Time) ([]plattformlogs.LogEntry, error) {
	query := `
        SELECT id, time, tag, data
        FROM fluentbit
        WHERE tag LIKE ? AND time < ?
        ORDER BY time DESC
        LIMIT ?;
    `
	resultList := make([]ScanResult, 0)
	tx := s.db.WithContext(ctx).Raw(query, track, t, pageSize).Scan(&resultList)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return convertScanResults(resultList)
}

func (s sqlLogNewStore) GetNewer(ctx context.Context, track string, t time.Time) ([]plattformlogs.LogEntry, error) {
	query := `
        SELECT id, time, tag, data
        FROM fluentbit
        WHERE tag = ? AND time >= ?
        ORDER BY time ASC
        LIMIT ?;
    `
	resultList := make([]ScanResult, 0)
	tx := s.db.WithContext(ctx).Raw(query, track, t, pageSize).Scan(&resultList)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return convertScanResults(resultList)
}

func (s sqlLogNewStore) GetNewerInstance(ctx context.Context, track string, t time.Time) ([]plattformlogs.LogEntry, error) {
	query := `
        SELECT id, time, tag, data
        FROM fluentbit
        WHERE tag LIKE ? AND time >= ?
        ORDER BY time ASC
        LIMIT ?;
    `
	resultList := make([]ScanResult, 0)
	tx := s.db.WithContext(ctx).Raw(query, track, t, pageSize).Scan(&resultList)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return convertScanResults(resultList)
}

func (s sqlLogNewStore) GetStartingIDUntilTime(ctx context.Context, track string, lastID int, t time.Time) ([]plattformlogs.LogEntry, error) {
	query := `
        SELECT id, time, tag, data
        FROM fluentbit
        WHERE tag = ? AND id >= ? time <= ?
        ORDER BY time ASC;
    `
	resultList := make([]ScanResult, 0)
	tx := s.db.WithContext(ctx).Raw(query, track, lastID, t).Scan(&resultList)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return convertScanResults(resultList)
}

func (s sqlLogNewStore) GetStartingIDUntilTimeInstance(ctx context.Context, track string, lastID int, t time.Time) ([]plattformlogs.LogEntry, error) {
	query := `
        SELECT id, time, tag, data
        FROM fluentbit
        WHERE tag LIKE ? AND id >= ? time <= ?
        ORDER BY time ASC;
    `
	resultList := make([]ScanResult, 0)
	tx := s.db.WithContext(ctx).Raw(query, track, lastID, t).Scan(&resultList)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return convertScanResults(resultList)
}

func convertScanResults(scanResults []ScanResult) ([]plattformlogs.LogEntry, error) {
	resultList := make([]plattformlogs.LogEntry, 0)

	for _, result := range scanResults {
		var dataMap map[string]interface{}
		err := json.Unmarshal(result.Data, &dataMap)
		if err != nil {
			return nil, err
		}

		resultList = append(resultList, plattformlogs.LogEntry{
			ID:   result.ID,
			Time: result.Time,
			Tag:  result.Tag,
			Data: dataMap,
		})
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
