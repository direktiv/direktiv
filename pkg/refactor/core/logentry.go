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
	LogTrackKey ContextKey = "track"
	LogTagsKey  ContextKey = "tags"
)

type LogEntry struct {
	ID   int
	Time time.Time
	Tag  string
	Data map[string]interface{}
}
