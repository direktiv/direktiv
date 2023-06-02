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
	cols = append(cols, "oid", "t", "level")
	vals = append(vals, uuid.New(), timestamp, level)
	cols = append(cols, "key")
	vals = append(vals, primaryKey)
	if len(keysAndValues) > 0 {
		b, err := json.Marshal(keysAndValues)
		if err != nil {
			return err
		}
		cols = append(cols, "entry")
		vals = append(vals, b)
	}
	secondaryKey, ok := keysAndValues["log_instance_call_path"]
	if ok {
		cols = append(cols, "secondary_key")
		vals = append(vals, secondaryKey)
	}

	q := "INSERT INTO log_msgs_v2 ("
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
	query := fmt.Sprintf("SELECT t, level, key, secondary_key, entry FROM log_msgs_v2 WHERE key='%v' ", primaryKey)

	prefix, ok := keysAndValues["log_instance_call_path"]
	if ok {
		query += fmt.Sprintf("AND secondary_key like '%s%%' ", prefix)
	}
	level, ok := keysAndValues["level"]
	if ok {
		query += fmt.Sprintf("AND level>='%v' ", level)
	}
	query += "ORDER BY t ASC "

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
		err := json.Unmarshal(e.Entry, &m)
		if err != nil {
			return nil, err
		}

		levels := []string{"debug", "info", "error"}
		m["level"] = levels[e.Level]
		msg := fmt.Sprintf("%v", m["msg"])
		delete(m, "msg")
		convertedList = append(convertedList, &logengine.LogEntry{
			T:      e.T,
			Msg:    msg,
			Fields: m,
		})
	}

	return convertedList, nil
}

type gormLogMsg struct {
	T            time.Time
	Level        int
	Key          string
	SecondaryKey string
	Entry        []byte
}
