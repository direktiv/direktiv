package core

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
)

const FlowCacheName = "flows"

const (
	FlowFileExtension        = ".wf.ts"
	FlowActionScopeWorkflow  = "workflow"
	FlowActionScopeNamespace = "namespace"
	FlowActionScopeSystem    = "system"
)

type Severity string

const (
	SeverityHint    Severity = "hint"
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

type ValidationError struct {
	Message     string   `json:"message"`
	StartLine   int      `json:"startLine"`
	StartColumn int      `json:"startColumn"`
	EndLine     int      `json:"endLine"`
	EndColumn   int      `json:"endColumn"`
	Severity    Severity `json:"severity"`
}

func (ve *ValidationError) Error() string {
	b, err := json.Marshal(ve)
	if err != nil {
		return fmt.Sprintf("%s (line: %d, column: %d)", ve.Message, ve.StartLine, ve.StartColumn)
	}

	return string(b)
}

type CompilerValidationError struct {
	Errors []*ValidationError
}

func (cve CompilerValidationError) Error() string {
	// return fmt.Sprintf("%s", strings.Join(cve.Errors, ", "))
	return ""
}

type ActionConfig struct {
	Type  string
	Cmd   string
	Size  string
	Image string
	Envs  []EnvironmentVariable

	Retries int

	// service file specific
	Scale int

	Auth *BasicAuthConfig
}
type BasicAuthConfig struct {
	Username string
	Password string
}
type StateView struct {
	Name        string   `json:"name"`
	Start       bool     `json:"start"`
	Finish      bool     `json:"finish"`
	Visited     bool     `json:"visited"`
	Failed      bool     `json:"failed"`
	Transitions []string `json:"transitions"`
}

type FlowConfig struct {
	Type       string
	Events     []EventConfig
	Cron       string
	Timeout    string
	State      string
	Actions    []ActionConfig
	Secrets    []string
	StateViews map[string]*StateView
}

type EventConfig struct {
	Type    string
	Context map[string]any
}

type TypescriptFlow struct {
	Script, Mapping string
	Config          FlowConfig
	Secrets         string // json map
}

// SortedStateViews returns the state views as a slice sorted by name.
func SortedStateViews(views map[string]*StateView) []*StateView {
	if views == nil {
		return nil
	}
	out := make([]*StateView, 0, len(views))
	for _, v := range views {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })

	return out
}

type Compiler interface {
	FetchScript(ctx context.Context, namespace, path string, withSecrets bool) (TypescriptFlow, error)
}
