package core

import (
	"time"
)

type LogStatus string

const (
	LogErrStatus       LogStatus = "error"
	LogUnknownStatus   LogStatus = "unknown"
	LogRunningStatus   LogStatus = "running"
	LogFailedStatus    LogStatus = "failed"
	LogCompletedStatus LogStatus = "completed"
)

type ContextKey string

const (
	// LogTrackKey is used for tracking related log entries, facilitating the organization of logs in sequences or chains.
	LogTrackKey ContextKey = "track"
	// LogTagsKey allows for adding structural metadata to log entries, for categorization based on their origin and context.
	LogTagsKey ContextKey = "tags"
)

type LogEntry struct {
	ID   int
	Time time.Time
	Tag  string
	Data map[string]interface{}
}
