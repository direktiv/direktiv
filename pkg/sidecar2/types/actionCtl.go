package types

import "github.com/direktiv/direktiv/pkg/refactor/engine"

type ActionController struct {
	engine.ActionRequest
	Cancel func()
}
