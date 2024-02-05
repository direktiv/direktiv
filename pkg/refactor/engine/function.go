package engine

import "github.com/google/uuid"

type FunctionContext struct {
	Timeout      int                   `json:"timeout"`
	InstanceID   uuid.UUID             `json:"instance_id"`
	State        string                `json:"state"`
	Step         int                   `json:"step"`
	Branch       int                   `json:"branch"`
	Callers      InstanceDescentInfo   `json:"callers"`
	Info         FunctionTelemetryInfo `json:"info"`
	WorkflowPath string                `json:"workflow_path"`
	AsyncExec    bool                  `json:"async_exec"`
}

type FunctionTelemetryInfo struct {
	InstanceTelemetryInfo
}
