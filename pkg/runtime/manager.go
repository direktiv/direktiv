package runtime

import (
	"github.com/direktiv/direktiv/pkg/compiler"
	"github.com/dop251/goja"
	"github.com/google/uuid"
)

type Manager interface {
	RuntimeData() *Data
	CreateInstance(id uuid.UUID, invoker, definition string) error
}

type Data struct {
	Program            *goja.Program
	Script             string
	Definition         *compiler.Definition
	Secrets, Functions map[string]string
}
