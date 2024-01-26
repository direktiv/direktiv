package engine

import "github.com/google/uuid"

type FunctionContext struct {
	Timeout      int
	InstanceID   uuid.UUID
	State        string
	Step         int
	Branch       int
	Callers      InstanceDescentInfo
	Info         FunctionTelemetryInfo
	WorkflowPath string
	AsyncExec    bool
}

type FunctionTelemetryInfo struct {
	InstanceTelemetryInfo
}
