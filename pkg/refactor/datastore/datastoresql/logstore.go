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

const (
	ns  = "namespace"
	wf  = "workflow"
	srv = "server"
	ins = "instance"
	mir = "mirror"
)

var _ logengine.LogStore = &sqlLogStore{} // Ensures SQLLogStore struct conforms to logengine.LogStore interface.

type sqlLogStore struct {
	db *gorm.DB
}

func (sl *sqlLogStore) Append(ctx context.Context, timestamp time.Time, level logengine.LogLevel, msg string, keysAndValues map[string]interface{}) error {
	cols := make([]string, 0, len(keysAndValues))
	vals := make([]interface{}, 0, len(keysAndValues))
	msg = strings.ReplaceAll(msg, "\u0000", "") // postgres will return an error if a string contains "\u0000"
	cols = append(cols, "oid", "t", "level", "msg")
	vals = append(vals, uuid.New(), timestamp, level, msg)
	databaseCols := []string{
		"instance_logs",
		"log_instance_call_path",
		"root_instance_id",
		"workflow_id",
		"namespace_logs",
		"mirror_activity_id",
	}
	for _, k := range databaseCols {
		if v, ok := keysAndValues[k]; ok {
			cols = append(cols, k)
			vals = append(vals, v)
		}
	}
	if len(keysAndValues) > 0 {
		b, err := json.Marshal(keysAndValues)
		if err != nil {
			return err
		}
		cols = append(cols, "tags")
		vals = append(vals, b)
	}
	q := "INSERT INTO log_msgs ("
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
	if keysAndValues["sender_type"] == srv {
		wEq = append(wEq, "workflow_id IS NULL")
		wEq = append(wEq, "namespace_logs IS NULL")
		wEq = append(wEq, "instance_logs IS NULL")
	}
	databaseCols := []string{
		"instance_logs",
		"root_instance_id",
		"workflow_id",
		"namespace_logs",
		"mirror_activity_id",
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
		err := json.Unmarshal(e.Tags, &m)
		if err != nil {
			return nil, err
		}

		levels := []string{"debug", "info", "error"}
		m["level"] = levels[e.Level]
		convertedList = append(convertedList, &logengine.LogEntry{
			T:      e.T,
			Msg:    e.Msg,
			Fields: m,
		})
	}

	return convertedList, nil
}

func composeQuery(limit, offset int, wEq []string) string {
	q := `SELECT t, msg, level, root_instance_id, log_instance_call_path, tags, workflow_id, mirror_activity_id, instance_logs, namespace_logs
	FROM log_msgs `
	q += "WHERE "
	for i, e := range wEq {
		q += e
		if i+1 < len(wEq) {
			q += " AND "
		}
	}
	q += " ORDER BY t ASC"
	if limit > 0 {
		q += fmt.Sprintf(" LIMIT %d ", limit)
	}
	if offset > 0 {
		q += fmt.Sprintf(" OFFSET %d ", offset)
	}

	return q + ";"
}

type gormLogMsg struct {
	T                   time.Time
	Msg                 string
	Level               int
	Tags                []byte
	WorkflowID          string
	MirrorActivityID    string
	InstanceLogs        string
	NamespaceLogs       string
	RootInstanceID      string
	LogInstanceCallPath string
}
