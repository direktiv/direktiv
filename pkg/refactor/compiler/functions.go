package compiler

import (
	"crypto/sha256"
	"fmt"
)

type Function struct {
	Image   string            `json:"image,omitempty"`
	Size    string            `json:"size"`
	Envs    map[string]string `json:"envs,omitempty"`
	Cmd     string            `json:"cmd,omitempty"`
	Init    []string          `json:"init,omitempty"`
	Flow    string            `json:"flow,omitempty"`
	Service string            `json:"service,omitempty"`
}

func (fn *Function) Validate() *Messages {

	m := newMessages()
	if fn.Image == "" && fn.Service == "" && fn.Flow == "" {
		m.addError("image, service or flow is required in function definition")
	}

	// TODO: check consistency of the function

	return m
}

func (fn *Function) GetID() string {
	str := fmt.Sprintf("%s-%s-%v-%s-%v-%s-%s",
		fn.Image, fn.Size, fn.Envs, fn.Cmd,
		fn.Init, fn.Flow, fn.Service)
	sh := sha256.Sum256([]byte(str))

	return fmt.Sprintf("fn-%x", sh[:8])
}

// GenerateFunctionID is used by the ruintime to generate the same id
func GenerateFunctionID(in interface{}) (string, error) {
	f, err := DoubleMarshal[Function](in)
	return f.GetID(), err
}
