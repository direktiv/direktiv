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

// LogLevel represents the severity level of logs in the system.
type LogLevel int

const (
	// LevelDebug is used for debug-level logs, providing detailed information.
	LevelDebug LogLevel = iota
	// LevelInfo is used for informational logs, indicating normal operation.
	LevelInfo
	// LevelWarn is used for warning-level logs, highlighting potential issues.
	LevelWarn
	// LevelError is used for error-level logs, indicating a failure or serious issue.
	LevelError
)

// String returns the string representation of the LogLevel for logging purposes.
func (level LogLevel) String() string {
	switch level {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "DEBUG"
	}
}
