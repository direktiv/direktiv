package core

import "context"

const FlowCacheName = "flows"

const (
	FlowFileExtension        = ".wf.ts"
	FlowActionScopeWorkflow  = "workflow"
	FlowActionScopeNamespace = "namespace"
	FlowActionScopeSystem    = "system"
)

type ActionConfig struct {
	Type  string
	Cmd   string
	Size  string
	Image string
	Envs  []EnvironmentVariable

	Retries int
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

type Compiler interface {
	FetchScript(ctx context.Context, namespace, path string, withSecrets bool) (TypescriptFlow, error)
}
