package instancestore

import (
	"time"
)

// This file contains structs and helper functions for those structs that will be used for data that gets marshalled to JSON before being inserted into the database.
// Probably all of these structs should be moved outside of this package over time.

// ParentInfo is part of the DescentInfo structure. It represents useful information about a single instance in the chain.
type ParentInfo struct {
	ID     string
	State  string
	Step   int
	Branch int // NOTE: renamed iterator to branch
}

// DescentInfo keeps a local copy of useful information about the entire chain of parent instances all the way to the root instance, excepting this instance.
type DescentInfo struct {
	Version string       // to let us identify and correct outdated versions of this struct
	Descent []ParentInfo // chain of callers from the root instance to the direct parent.
}

// ChildInfo is part of the ChildrenInfo structure. It represents useful information about a single child action.
type ChildInfo struct {
	ID          string
	Async       bool
	Complete    bool
	Type        string
	Attempts    int
	ServiceName string
}

// ChildrenInfo keeps some useful information about all direct child actions of this instance.
type ChildrenInfo struct {
	Version  string // to let us identify and correct outdated versions of this struct
	Children []ChildInfo
}

// RuntimeInfo keeps other miscellaneous information useful to the engine.
type RuntimeInfo struct {
	Version        string // to let us identify and correct outdated versions of this struct
	Controller     string
	Flow           []string // NOTE: now that we keep a copy of the definition we could replace []string with []int
	StateBeginTime time.Time
	Attempts       int
}

// Settings keeps a local copy of various namespace and workflow settings so that the engine doesn't have to look them up separately.
type Settings struct {
	Version     string // to let us identify and correct outdated versions of this struct
	LogToEvents string
	// NOTE: only one field for now, but soon this could take a copy of namespace event settings, as well as whether to print debug logs
}

// TelemetryInfo keeps information useful to our telemetry logic.
type TelemetryInfo struct {
	Version  string // to let us identify and correct outdated versions of this struct
	TraceID  string
	SpanID   string
	CallPath string
}
