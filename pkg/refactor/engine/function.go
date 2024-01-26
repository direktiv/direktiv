package engine

import "github.com/google/uuid"

type ExecContextKeyType string

const (
	ExecContextKey ExecContextKeyType = "ExecContext"
)

type FunctionContext struct {
	Timeout      int
	InstanceID   uuid.UUID
	State        string
	Step         int // TODO: remove me
	Branch       int // TODO: remove me
	Callers      InstanceDescentInfo
	Info         FunctionTelemetryInfo
	WorkflowPath string
	AsyncExec    bool
}

type FunctionTelemetryInfo struct {
	InstanceTelemetryInfo
}
