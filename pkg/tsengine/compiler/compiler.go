package compiler

import (
	"crypto/sha256"
	"fmt"
	"regexp"
)

type Compiler struct {
	JavaScript string
}

func New(path, typeScript string) (*Compiler, error) {
	return &Compiler{}, nil
}

func (c *Compiler) CompileFlow() (*FlowInformation, error) {
	flowInformation := &FlowInformation{
		Definition: DefaultDefinition(),
		Functions:  make(map[string]Function),
		ID:         c.getID(),
	}

	return flowInformation, nil
}

func (c *Compiler) getID() string {

	str := fmt.Sprintf("%s-%s", "TODO: NAMESPACE", c.JavaScript)
	sh := sha256.Sum256([]byte(str))

	whitelist := regexp.MustCompile("[^a-zA-Z0-9]+")
	str = whitelist.ReplaceAllString(str, "-")

	// Prevent too long ids
	if len(str) > 50 {
		str = str[:50]
	}

	return fmt.Sprintf("%s-%x", str, sh[:5])
}

type FlowInformation struct {
	ID         string
	Definition *Definition

	Functions map[string]Function
}

type Function struct {
	Image   string            `json:"image,omitempty"`
	Size    string            `json:"size"`
	Envs    map[string]string `json:"envs,omitempty"`
	Cmd     string            `json:"cmd,omitempty"`
	Init    []string          `json:"init,omitempty"`
	Flow    string            `json:"flow,omitempty"`
	Service string            `json:"service,omitempty"`
}
