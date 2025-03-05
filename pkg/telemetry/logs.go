package telemetry

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/direktiv/direktiv/pkg/core"
)

type LogLevel int

type InstanceInfo struct {
	Namespace string         `json:"namespace"`
	Instance  string         `json:"instance"`
	Invoker   string         `json:"invoker"`
	Callpath  string         `json:"callpath"`
	Path      string         `json:"path"`
	State     string         `json:"state"`
	Status    core.LogStatus `json:"status"`
}

const (
	LogLevelDebug LogLevel = iota
	LogLevelWarn
	LogLevelInfo
	LogLevelError

	DirektivInstance
)

func InitInstanceLog(ctx context.Context, info InstanceInfo) context.Context {

	fmt.Printf("INSTANCE %+v\n", info)

	// check if required fields are set
	return context.WithValue(ctx, DirektivInstance, info)
}

func LogInstance(ctx context.Context, lvl LogLevel, msg string) {
	switch lvl {
	case LogLevelDebug:
		slog.DebugContext(ctx, msg)
	case LogLevelWarn:
		slog.WarnContext(ctx, msg)
	case LogLevelInfo:
		slog.InfoContext(ctx, msg)
	case LogLevelError:
		slog.ErrorContext(ctx, msg)
	}
}
