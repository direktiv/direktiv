package core

import "context"

type FlowConfig struct {
	Type    string
	Events  []*EventConfig
	Cron    string
	Timeout string
	State   string
}

type EventConfig struct {
	Type    string
	Context map[string]any
}

type TypescriptFlow struct {
	Script, Mapping string
	Config          FlowConfig
}

type Compiler interface {
	FetchScript(ctx context.Context, namespace, path string) (*TypescriptFlow, error)
}
