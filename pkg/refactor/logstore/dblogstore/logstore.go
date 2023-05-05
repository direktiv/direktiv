package dblogstore

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/logstore"
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

var _ logstore.LogStore = &SQLLogStore{} // Ensures SQLLogStore struct conforms to logstore.LogStore interface.

func NewSQLLogStore(db *gorm.DB) logstore.LogStore {
	return &SQLLogStore{
		db: db,
	}
}

type SQLLogStore struct {
	db *gorm.DB
}

// Append implements logstore.LogStore.
// For instance-logs following Key Value pairs SHOULD be present: instance-id, callpath, roottinsanceid
// For namespace-logs following Key Value pairs SHOULD be present: namespace-id
// For mirror-logs following Key Value pairs SHOULD be present: mirror-id
// For workflow-logs following Key Value pairs SHOULD be present: workflow_id
// Any other keysAndValues pair will be stored as tags attached to the log-entry.
func (sl *SQLLogStore) Append(ctx context.Context, timestamp time.Time, msg string, keysAndValues ...interface{}) error {
	fields, err := mapKeysAndValues(keysAndValues...)
	cols := make([]string, 0, len(keysAndValues))
	vals := make([]interface{}, 0, len(keysAndValues))
	if err != nil {
		return err
	}
	lvl, err := getLevel(fields)
	if err != nil {
		return err
	}
	delete(fields, "level")
	cols = append(cols, "t", "msg", "level")
	vals = append(vals, timestamp, msg, lvl)
	value, ok := fields["instance-id"]
	if ok {
		cols = append(cols, "instance_logs")
		vals = append(vals, value)
	}
	value, ok = fields["callpath"]
	if ok {
		cols = append(cols, "log_instance_call_path")
		vals = append(vals, value)
		delete(fields, "callpath")
	}
	value, ok = fields["rootInstanceID"]
	if ok {
		cols = append(cols, "root_instance_id")
		vals = append(vals, value)
		delete(fields, "rootInstanceID")
	}
	value, ok = fields["workflow-id"]
	if ok {
		cols = append(cols, "workflow_id")
		vals = append(vals, value)
	}
	value, ok = fields["namespace-id"]
	if ok {
		cols = append(cols, "namespace_logs")
		vals = append(vals, value)
	}
	value, ok = fields["mirror-id"]
	if ok {
		cols = append(cols, "mirror_activity_id")
		vals = append(vals, value)
	}
	b, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	cols = append(cols, "tags")
	vals = append(vals, b)
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

// Get implements logstore.LogStore.
// To query server-logs pass: "recipientType", "server" via keysAndValues
// This method will search for any of those keys and query all logs:
// level, workflow-id, namespace-id, callpath, rootInstanceID-id, mirror-activity-id, limit, offset
// limit, offset MUST be passed as integer and are useful for pagination.
// level SHOULD be passed as a string. Valid values are "debug", "info", "error", "panic". Returned log-entries will have same or higher level as the passed one.
// passing a callpath will return all logs which have a callpath with the prefix as the passed callpath value.
// when passing callpath the rootInstanceID-id SHOULD be passed to optimize the performance of the query.
func (sl *SQLLogStore) Get(ctx context.Context, keysAndValues ...interface{}) ([]*logstore.LogEntry, error) {
	fields, err := mapKeysAndValues(keysAndValues...)
	if err != nil {
		return nil, err
	}
	wEq := []string{}
	if fields["recipientType"] == srv {
		wEq = append(wEq, "workflow_id IS NULL")
		wEq = append(wEq, "namespace_logs IS NULL")
		wEq = append(wEq, "instance_logs IS NULL")
	}
	level, ok := fields["level"]
	if ok {
		levelValue, ok := level.(string)
		if !ok {
			return nil, fmt.Errorf("level should be a string")
		}
		wEq = setMinumLogLevel(levelValue, wEq)
	}
	id, ok := fields["workflow-id"]
	if ok {
		wEq = append(wEq, fmt.Sprintf("workflow_id = '%s'", id))
	}
	id, ok = fields["namespace-id"]
	if ok {
		wEq = append(wEq, fmt.Sprintf("namespace_logs = '%s'", id))
	}
	prefix, ok := fields["callpath"]
	if ok {
		wEq = append(wEq, fmt.Sprintf("log_instance_call_path like '%s%%'", prefix))
	}
	id, ok = fields["rootInstanceID-id"]
	if ok {
		wEq = append(wEq, fmt.Sprintf("root_instance_id = '%s'", id))
	}
	id, ok = fields["mirror-activity-id"]
	if ok {
		wEq = append(wEq, fmt.Sprintf("mirror_activity_id = '%s'", id))
	}
	limit, ok := fields["limit"]
	var limitValue int
	if ok {
		limitValue, ok = limit.(int)
		if !ok {
			return nil, fmt.Errorf("limit should be an int")
		}
	}
	offset, ok := fields["offset"]
	var offsetValue int
	if ok {
		offsetValue, ok = offset.(int)
		if !ok {
			return nil, fmt.Errorf("offset should be an int")
		}
	}
	query := composeQuery(limitValue, offsetValue, wEq)

	resultList := make([]*gormLogMsg, 0)
	tx := sl.db.WithContext(ctx).Raw(query).Scan(&resultList)
	if tx.Error != nil {
		return nil, tx.Error
	}
	convertedList := make([]*logstore.LogEntry, 0, len(resultList))
	for _, e := range resultList {
		m := make(map[string]interface{})
		for k, e := range e.Tags {
			m[k] = e
		}
		m["level"] = e.Level
		convertedList = append(convertedList, &logstore.LogEntry{
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
	if limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		q += fmt.Sprintf(" OFFSET %d", offset)
	}

	return q
}

func getLevel(fields map[string]interface{}) (string, error) {
	lvl, ok := fields["level"]
	if !ok {
		return "", fmt.Errorf("level was missing as argument in the keysAndValues pair")
	}
	switch lvl {
	case "debug", "info", "error", "panic":
		return fmt.Sprintf("%s", lvl), nil
	}

	return "", fmt.Errorf("field level provided in keysAndValues has an invalid value %s", lvl)
}

func mapKeysAndValues(keysAndValues ...interface{}) (map[string]interface{}, error) {
	fields := make(map[string]interface{})
	if len(keysAndValues) == 0 || len(keysAndValues)%2 != 0 {
		return nil, fmt.Errorf("keysAndValues was not a list of key value pairs")
	}
	for i := 0; i < len(keysAndValues); i += 2 {
		key := fmt.Sprintf("%s", keysAndValues[i])
		fields[key] = keysAndValues[i+1]
	}

	return fields, nil
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
