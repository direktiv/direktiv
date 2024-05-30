package datastore

import (
	"context"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/core"
)

type LogLevel int

const (
	Debug LogLevel = iota
	Info
	Warn
	Error
)

// LogStore manages storing and querying LogEntries.
type LogStore interface {
	GetNewer(ctx context.Context, track string, t time.Time) ([]core.LogEntry, error)
	GetOlder(ctx context.Context, track string, t time.Time) ([]core.LogEntry, error)
	GetStartingIDUntilTime(ctx context.Context, track string, lastID int, t time.Time) ([]core.LogEntry, error)
	GetNewerInstance(ctx context.Context, track string, t time.Time) ([]core.LogEntry, error)
	GetOlderInstance(ctx context.Context, track string, t time.Time) ([]core.LogEntry, error)
	GetStartingIDUntilTimeInstance(ctx context.Context, track string, lastID int, t time.Time) ([]core.LogEntry, error)
	DeleteOldLogs(ctx context.Context, t time.Time) error
}
