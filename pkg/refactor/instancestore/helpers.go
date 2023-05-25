package instancestore

import (
	"time"
)

// This file contains structs and helper functions for those structs that will be used for data that gets marshalled to JSON before being inserted into the database.

// TODO: alan, move this outside this package?
type ParentInfo struct {
	ID     string
	State  string
	Step   int
	Branch int // NOTE: renamed iterator to branch
}

type DescentInfo struct {
	Version string       // to let us identify and correct outdated versions of this struct
	Descent []ParentInfo // Chain of callers from the root instance to the direct parent.
}

// TODO: alan, move this outside this package?
type ChildInfo struct {
	ID          string
	Async       bool
	Complete    bool
	Type        string
	Attempts    int
	ServiceName string
}

type ChildrenInfo struct {
	Version  string // to let us identify and correct outdated versions of this struct
	Children []ChildInfo
}

// TODO: alan, move this outside this package?
type RuntimeInfo struct {
	Version        string // to let us identify and correct outdated versions of this struct
	Controller     string
	Flow           []string // TODO: alan, now that we keep a copy of the definition we can replace []string with []int
	StateBeginTime time.Time
	Attempts       int
}

// TODO: alan, move this outside this package?
type Settings struct {
	Version     string // to let us identify and correct outdated versions of this struct
	LogToEvents string
	// NOTE: only one field for now, but soon this could take a copy of namespace event settings, as well as whether to print debug logs
}

// TODO: alan, move this outside this package?
type TelemetryInfo struct {
	Version string // to let us identify and correct outdated versions of this struct
	TraceID string
	SpanID  string
}
