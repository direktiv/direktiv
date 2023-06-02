package datastoresql

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/logengine"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var _ logengine.LogStore = &sqlLogStore{} // Ensures SQLLogStore struct conforms to logengine.LogStore interface.

type sqlLogStore struct {
	db *gorm.DB
}

func (sl *sqlLogStore) Append(ctx context.Context, timestamp time.Time, level logengine.LogLevel, msg string, primaryKey string, keysAndValues map[string]interface{}) error {
	cols := make([]string, 0, len(keysAndValues))
	vals := make([]interface{}, 0, len(keysAndValues))
	msg = strings.ReplaceAll(msg, "\u0000", "") // postgres will return an error if a string contains "\u0000"
	keysAndValues["msg"] = msg
	cols = append(cols, "id", "timestamp", "log_level")
	vals = append(vals, uuid.New(), timestamp, level)
	cols = append(cols, "primary_key")
	vals = append(vals, primaryKey)
	b, err := json.Marshal(keysAndValues)
	if err != nil {
		return err
	}
	cols = append(cols, "log_entry")
	vals = append(vals, b)
	secondaryKey, ok := keysAndValues["log_instance_call_path"]
	if ok {
		cols = append(cols, "secondary_key")
		vals = append(vals, secondaryKey)
	}

	q := "INSERT INTO log_entries ("
	qTail := "VALUES ("
	for i := range vals {
		q += fmt.Sprintf(cols[i])
		qTail += fmt.Sprintf("$%d", i+1)
		if i != len(vals)-1 {
			q += ", "
			qTail += ", "
		}
	}
	q = q + ") " + qTail + ");"
	tx := sl.db.WithContext(ctx).Exec(q, vals...)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return fmt.Errorf("no rows affected")
	}

	return nil
}

func (sl *sqlLogStore) Get(ctx context.Context, limit, offset int, primaryKey string, keysAndValues map[string]interface{}) ([]*logengine.LogEntry, error) {
	query := fmt.Sprintf("SELECT timestamp, log_level, primary_key, secondary_key, log_entry FROM log_entries WHERE primary_key='%v' ", primaryKey)

	prefix, ok := keysAndValues["log_instance_call_path"]
	if ok {
		query += fmt.Sprintf("AND secondary_key like '%s%%' ", prefix)
	}
	level, ok := keysAndValues["level"]
	if ok {
		query += fmt.Sprintf("AND log_level>='%v' ", level)
	}
	query += "ORDER BY timestamp ASC "

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d ", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d ", offset)
	}
	query += ";"
	resultList := make([]*gormLogMsg, 0)
	tx := sl.db.WithContext(ctx).Raw(query).Scan(&resultList)
	if tx.Error != nil {
		return nil, tx.Error
	}
	convertedList := make([]*logengine.LogEntry, 0, len(resultList))
	for _, e := range resultList {
		m := make(map[string]interface{})
		err := json.Unmarshal(e.LogEntry, &m)
		if err != nil {
			return nil, err
		}

		levels := []string{"debug", "info", "error"}
		m["level"] = levels[e.LogLevel]
		msg := fmt.Sprintf("%v", m["msg"])
		delete(m, "msg")
		convertedList = append(convertedList, &logengine.LogEntry{
			T:      e.Timestamp,
			Msg:    msg,
			Fields: m,
		})
	}

	return convertedList, nil
}

type gormLogMsg struct {
	Timestamp    time.Time
	LogLevel     int
	PrimaryKey   string
	SecondaryKey string
	LogEntry     []byte
}
