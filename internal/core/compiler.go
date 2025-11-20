package core

import "context"

const FlowCacheName = "flows"

const (
	FlowFileExtension        = ".wf.ts"
	FlowActionScopeLocal     = "local"
	FlowActionScopeNamespace = "namespace"
	FlowActionScopeSystem    = "system"
)

type ActionConfig struct {
	Type  string
	Cmd   string
	Size  string
	Image string
	Envs  []EnvironmentVariable

	Service string

	Retries int
	// Patches []ServicePatch
}

type FlowConfig struct {
	Type    string
	Events  []EventConfig
	Cron    string
	Timeout string
	State   string
	Actions []ActionConfig
	Secrets []string
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
