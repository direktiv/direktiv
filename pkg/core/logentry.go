package core

import (
	"time"
)

type LogStatus string

const (
	LogErrStatus       LogStatus = "error"   // deprecated
	LogUnknownStatus   LogStatus = "unknown" // deprecated
	LogRunningStatus   LogStatus = "running"
	LogFailedStatus    LogStatus = "failed" // deprecated
	LogCompletedStatus LogStatus = "completed"
)

type ContextKey string

const (
	// LogTrackKey is used for tracking related log entries, facilitating the organization of logs in sequences or chains.
	LogTrackKey ContextKey = "track"
)

type LogEntry struct {
	ID   int
	Time time.Time
	Tag  string
	Data map[string]interface{}
}
