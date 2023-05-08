package sql

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

// Append implements logengine.LogStore.
// - For instance-logs following Key Value pairs SHOULD be present: instance_logs, log_instance_call_path, root_instance_id
// - For namespace-logs following Key Value pairs SHOULD be present: namespace_logs
// - For mirror-logs following Key Value pairs SHOULD be present: mirror_activity_id
// - For workflow-logs following Key Value pairs SHOULD be present: workflow_id
// - All passed keysAndValues pair will be stored attached to the log-entry.
func (sl *sqlLogStore) Append(ctx context.Context, timestamp time.Time, msg string, keysAndValues map[string]interface{}) error {
	cols := make([]string, 0, len(keysAndValues))
	vals := make([]interface{}, 0, len(keysAndValues))
	msg = strings.ReplaceAll(msg, "\u0000", "") // postgres will return an error if a string contains "\u0000"
	cols = append(cols, "oid", "t", "msg")
	vals = append(vals, uuid.New(), timestamp, msg)
	databaseCols := []string{
		"instance_logs",
		"log_instance_call_path",
		"root_instance_id",
		"workflow_id",
		"namespace_logs",
		"mirror_activity_id",
		"level",
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
		q += fmt.Sprintf("'%s'", cols[i])
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

// Get implements logengine.LogStore.
// - To query server-logs pass: "recipientType", "server" via keysAndValues
// - level SHOULD be passed as a string. Valid values are "debug", "info", "error", "panic".
// - This method will search for any of followings keys and query all matching logs:
// level, workflow_id, namespace_logs, log_instance_call_path, root_instance_id, mirror_activity_id
// Any other not mentioned passed key value pair will be ignored.
// Returned log-entries will have same or higher level as the passed one.
// - Passing a log_instance_call_path will return all logs which have a callpath with the prefix as the passed log_instance_call_path value.
// when passing log_instance_call_path the root_instance_id SHOULD be passed to optimize the performance of the query.
func (sl *sqlLogStore) Get(ctx context.Context, keysAndValues map[string]interface{}, limit, offset int) ([]*logengine.LogEntry, error) {
	wEq := []string{}
	if keysAndValues["recipientType"] == srv {
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
		levelValue, ok := level.(string)
		if !ok {
			return nil, fmt.Errorf("level should be a string")
		}
		wEq = setMinumLogLevel(levelValue, wEq)
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
		for k, e := range e.Tags {
			m[k] = e
		}
		m["level"] = e.Level
		convertedList = append(convertedList, &logengine.LogEntry{
			T:      e.T,
			Msg:    e.Msg,
			Fields: m,
		})
	}

	return convertedList, nil
}

func composeQuery(limit, offset int, wEq []string) string {
	q := `SELECT oid, t, msg, level, root_instance_id, log_instance_call_path, tags, workflow_id, mirror_activity_id, instance_logs, namespace_logs
	FROM log_msgs `
	q += "WHERE "
	for i, e := range wEq {
		q += e
		if i+1 < len(wEq) {
			q += " AND "
		}
	}
	q += " ORDER BY t ASC"
	if limit >= 0 {
		q += fmt.Sprintf(" LIMIT %d ", limit)
	}
	if offset >= 0 {
		q += fmt.Sprintf(" OFFSET %d ", offset)
	}

	return q + ";"
}

type gormLogMsg struct {
	Oid                 uuid.UUID
	T                   time.Time
	Msg                 string
	Level               string
	Tags                map[string]interface{}
	WorkflowID          string
	MirrorActivityID    string
	InstanceLogs        string
	NamespaceLogs       string
	RootInstanceID      string
	LogInstanceCallPath string
}

func setMinumLogLevel(loglevel string, wEq []string) []string {
	levels := []string{"debug", "info", "error", "panic"}
	switch loglevel {
	case "debug":
		return wEq
	case "info":
		levels = levels[1:]
	case "error":
		levels = levels[2:]
	case "panic":
		levels = levels[3:]
	}
	q := "( "
	for i, e := range levels {
		q += fmt.Sprintf("level = '%s' ", e)
		if i < len(levels)-1 {
			q += " OR "
		}
	}
	q += " )"

	return append(wEq, q)
}
