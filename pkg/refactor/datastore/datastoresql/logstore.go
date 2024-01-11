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
	valuesCopy := make(map[string]interface{})
	for k, v := range keysAndValues {
		valuesCopy[k] = v
	}

	msg = strings.ReplaceAll(msg, "\u0000", "")
	valuesCopy["message"] = msg

	topic := fmt.Sprintf("%v%v%v", valuesCopy["type"], valuesCopy["root_instance_id"], time.Now().UTC().Format("2006-01-02"))

	query := `
		INSERT INTO engine_messages 
		(id, timestamp, level, topic, source, entry) 
		VALUES (?, ?, ?, ?, ?, ?)
	`

	vals := []interface{}{
		uuid.New(),
		timestamp,
		level,
		topic,
		valuesCopy["source"],
	}

	b, err := json.Marshal(valuesCopy)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	vals = append(vals, b)

	tx := sl.db.WithContext(ctx).Exec(query, vals...)
	if tx.Error != nil {
		return fmt.Errorf("failed to execute SQL query: %w", tx.Error)
	}
	if tx.RowsAffected == 0 {
		return fmt.Errorf("no rows affected by the SQL query")
	}

	return nil
}

func (sl *sqlLogStore) Get(ctx context.Context, keysAndValues map[string]interface{}, limit, offset int) ([]*logengine.LogEntry, int, error) {
	today := fmt.Sprintf("%v%v%v", keysAndValues["type"], keysAndValues["root_instance_id"], time.Now().UTC().Format("2006-01-02"))
	yesterday := fmt.Sprintf("%v%v%v", keysAndValues["type"], keysAndValues["root_instance_id"], time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02"))

	delete(keysAndValues, "type")
	delete(keysAndValues, "root_instance_id")

	wEq := []string{
		fmt.Sprintf(`topic = '%s' OR topic = '%s'`, today, yesterday), // we query only the logs for the last 2 days
	}

	addCondition(&wEq, "source", keysAndValues)
	addCondition(&wEq, "level", keysAndValues)
	addConditionPrefix(&wEq, "log_instance_call_path", keysAndValues)

	query := buildQuery("timestamp, level, log_instance_call_path, source, entry", wEq, limit, offset)

	resultList := make([]*gormLogMsg, 0)
	tx := sl.db.WithContext(ctx).Raw(query).Scan(&resultList)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}

	countQuery := buildCountQuery(wEq)
	count := 0
	tx = sl.db.WithContext(ctx).Raw(countQuery).Scan(&count)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}

	convertedList := make([]*logengine.LogEntry, 0, len(resultList))
	for _, e := range resultList {
		m := make(map[string]interface{})
		err := json.Unmarshal(e.Entry, &m)
		if err != nil {
			return nil, 0, err
		}

		m["level"] = getLevelString(e.Level)
		msg := fmt.Sprintf("%v", m["message"])
		delete(m, "message")
		convertedList = append(convertedList, &logengine.LogEntry{
			T:      e.Timestamp,
			Msg:    msg,
			Fields: m,
		})
	}

	return convertedList, count, nil
}

func addCondition(wEq *[]string, key string, keysAndValues map[string]interface{}) {
	if val, ok := keysAndValues[key]; ok {
		*wEq = append(*wEq, fmt.Sprintf("%s = '%s'", key, val))
	}
}

func addConditionPrefix(wEq *[]string, key string, keysAndValues map[string]interface{}) {
	if val, ok := keysAndValues[key]; ok {
		*wEq = append(*wEq, fmt.Sprintf("%s LIKE '%s%%'", key, val))
	}
}

func buildQuery(fields string, wEq []string, limit, offset int) string {
	query := fmt.Sprintf(`
		SELECT %s
		FROM engine_messages
		WHERE %s
		ORDER BY timestamp ASC`, fields, strings.Join(wEq, " AND "))

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	return query
}

func buildCountQuery(wEq []string) string {
	return fmt.Sprintf(`
		SELECT COUNT(id)
		FROM engine_messages
		WHERE %s`, strings.Join(wEq, " AND "))
}

func getLevelString(level int) string {
	levels := []string{"debug", "info", "warn", "error"}
	if level >= 0 && level < len(levels) {
		return levels[level]
	}

	return "debug"
}

func (sl *sqlLogStore) DeleteOldLogs(ctx context.Context, t time.Time) error {
	query := "DELETE FROM engine_messages WHERE timestamp < $1"

	res := sl.db.WithContext(ctx).Exec(
		query,
		t.UTC(),
	)
	if res.Error != nil {
		return res.Error
	}

	return nil
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
