package core

type VariableScope string

const (
	VariableScopeNamespace VariableScope = "namespace"
	VariableScopeWorkflow  VariableScope = "workflow"
	VariableScopeInstance  VariableScope = "instance"
)
