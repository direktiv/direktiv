package logengine

import (
	"context"
	"time"
)

// LogStore manages storing and querying LogEntries.
type LogStore interface {
	// appends a log entry to the logs. Passed keysAnValues will be associated with the log entry.
	// - For instance-logs following Key Value pairs SHOULD be present: instance_logs, log_instance_call_path, root_instance_id
	// - For namespace-logs following Key Value pairs SHOULD be present: namespace_logs
	// - For mirror-logs following Key Value pairs SHOULD be present: mirror_activity_id
	// - For workflow-logs following Key Value pairs SHOULD be present: workflow_id
	// - All passed keysAndValues pair will be stored attached to the log-entry.
	Append(ctx context.Context, timestamp time.Time, level LogLevel, msg string, keysAndValues map[string]interface{}) error
	// returns a limited number of log-entries that have matching associated fields with the provided keysAndValues pairs
	// starting a given offset. For no offset or unlimited log-entries in the result set the value to 0.
	// - To query server-logs pass: "sender_type", "server" via keysAndValues
	// - level SHOULD be passed as a string. Valid values are "debug", "info", "error", "panic".
	// - This method will search for any of followings keys and query all matching logs:
	// level, workflow_id, namespace_logs, log_instance_call_path, root_instance_id, mirror_activity_id
	// Any other not mentioned passed key value pair will be ignored.
	// Returned log-entries will have same or higher level as the passed one.
	// - Passing a log_instance_call_path will return all logs which have a callpath with the prefix as the passed log_instance_call_path value.
	// when passing log_instance_call_path the root_instance_id SHOULD be passed to optimize the performance of the query.
	Get(ctx context.Context, keysAndValues map[string]interface{}, limit, offset int) ([]*LogEntry, error)
}

// Represents an individual log entry.
type LogEntry struct {
	// the timestamp of the log-entry.
	T   time.Time
	Msg string
	// Fields contains metadata of the log-entry.
	Fields map[string]interface{}
}

type LogLevel int

const (
	Debug LogLevel = iota
	Info           = iota
	Error          = iota
)

func FilterLogs(logs []*LogEntry, keysAndValues map[string]interface{}) []*LogEntry {
	databaseCols := []string{
		"instance_logs",
		"log_instance_call_path",
		"root_instance_id",
		"workflow_id",
		"namespace_logs",
		"mirror_activity_id",
	}

	for k := range keysAndValues {
		for _, v2 := range databaseCols {
			if v2 == k {
				delete(keysAndValues, k)
			}
		}
	}

	filteredLogs := make([]*LogEntry, 0)

	for _, l := range logs {
		if shouldAdd(keysAndValues, l.Fields) {
			filteredLogs = append(filteredLogs, l)
		}
	}

	return filteredLogs
}

// returns true if all key values pairs are present in the fields and the values match
// returns always true if keyAndValues is empty
func shouldAdd(keysAndValues map[string]interface{}, fields map[string]interface{}) bool {
	match := true
	for k, e := range keysAndValues {
		t, ok := fields[k]
		if ok {
			match = match && e == t
		}
	}
	return match
}
