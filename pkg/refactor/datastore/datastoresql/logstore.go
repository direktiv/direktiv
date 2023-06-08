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

func (sl *sqlLogStore) Append(ctx context.Context, timestamp time.Time, level logengine.LogLevel, msg string, keysAndValues map[string]interface{}) error {
	cols := make([]string, 0, len(keysAndValues))
	vals := make([]interface{}, 0, len(keysAndValues))
	msg = strings.ReplaceAll(msg, "\u0000", "") // postgres will return an error if a string contains "\u0000"
	cols = append(cols, "oid", "timestamp", "level")
	vals = append(vals, uuid.New(), timestamp, level)
	databaseCols := []string{
		"source",
		"log_instance_call_path",
		"root_instance_id",
		"type",
	}
	for _, k := range databaseCols {
		if v, ok := keysAndValues[k]; ok {
			cols = append(cols, k)
			vals = append(vals, v)
		}
	}
	keysAndValues["message"] = msg
	b, err := json.Marshal(keysAndValues)
	if err != nil {
		return err
	}
	cols = append(cols, "entry")
	vals = append(vals, b)
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

func (sl *sqlLogStore) Get(ctx context.Context, keysAndValues map[string]interface{}, limit, offset int) ([]*logengine.LogEntry, error) {
	wEq := []string{}

	databaseCols := []string{
		"source",
		"type",
		"root_instance_id",
	}
	for _, k := range databaseCols {
		if v, ok := keysAndValues[k]; ok {
			wEq = append(wEq, fmt.Sprintf("%s = '%s'", k, v))
		}
	}
	level, ok := keysAndValues["level"]
	if ok {
		wEq = append(wEq, fmt.Sprintf("level >= '%v' ", level))
	}
	prefix, ok := keysAndValues["log_instance_call_path"]
	if ok {
		wEq = append(wEq, fmt.Sprintf("log_instance_call_path like '%s%%'", prefix))
	}

	query := composeQuery(limit, offset, wEq)

	resultList := make([]*gormLogMsg, 0)
	tx := sl.db.WithContext(ctx).Raw(query).Scan(&resultList)
	if tx.Error != nil {
		return nil, tx.Error
	}
	convertedList := make([]*logengine.LogEntry, 0, len(resultList))
	for _, e := range resultList {
		m := make(map[string]interface{})
		err := json.Unmarshal(e.Entry, &m)
		if err != nil {
			return nil, err
		}

		levels := []string{"debug", "info", "error"}
		m["level"] = levels[e.Level]
		msg := fmt.Sprintf("%v", m["message"])
		delete(m, "message")
		convertedList = append(convertedList, &logengine.LogEntry{
			T:      e.Timestamp,
			Msg:    msg,
			Fields: m,
		})
	}

	return convertedList, nil
}

func composeQuery(limit, offset int, wEq []string) string {
	q := `SELECT timestamp, level, root_instance_id, log_instance_call_path, source, type, entry
	FROM log_entries `
	q += "WHERE "
	for i, e := range wEq {
		q += e
		if i+1 < len(wEq) {
			q += " AND "
		}
	}
	q += " ORDER BY timestamp ASC"
	if limit > 0 {
		q += fmt.Sprintf(" LIMIT %d ", limit)
	}
	if offset > 0 {
		q += fmt.Sprintf(" OFFSET %d ", offset)
	}

	return q + ";"
}

type gormLogMsg struct {
	Timestamp           time.Time
	Level               int
	Entry               []byte
	Source              uuid.UUID
	Type                string
	RootInstanceID      string
	LogInstanceCallPath string
}
