package core

import "context"

const (
	FlowFileExtension        = ".wf.ts"
	FlowActionScopeLocal     = "local"
	FlowActionScopeNamespace = "namespace"
	FlowActionScopeSystem    = "system"
)

type ActionConfig struct {
	Type  string
	Size  string
	Image string
	Envs  map[string]string
}

type FlowConfig struct {
	Type    string
	Events  []*EventConfig
	Cron    string
	Timeout string
	State   string
	Actions []ActionConfig
}

type EventConfig struct {
	Type    string
	Context map[string]any
}

type TypescriptFlow struct {
	Script, Mapping string
	Config          *FlowConfig
}

type Compiler interface {
	FetchScript(ctx context.Context, namespace, path string) (*TypescriptFlow, error)
}
