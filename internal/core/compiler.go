package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

const (
	FlowFileExtension        = ".wf.ts"
	FlowActionScopeLocal     = "local"
	FlowActionScopeNamespace = "namespace"
	FlowActionScopeSystem    = "system"
	FlowActionScopeSubflow   = "subflow"
)

type ActionConfig struct {
	Type   string
	Size   string
	Image  string
	Inject bool
	Envs   map[string]string
}

func (ac *ActionConfig) ID(svcType, namespace, path string) (string, error) {
	type genID struct {
		ac              ActionConfig
		namespace, path string
	}

	a := genID{
		ac:        *ac,
		namespace: namespace,
		path:      path,
	}

	j, err := json.Marshal(a)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(j)

	return fmt.Sprintf("%s-%s", svcType, hex.EncodeToString(hash[:32])), nil
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
	Namespace, Path string
	Script, Mapping string
	Config          *FlowConfig
}

type Compiler interface {
	FetchScript(ctx context.Context, namespace, path string) (*TypescriptFlow, error)
}
