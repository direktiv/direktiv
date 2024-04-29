package flow

import (
	"context"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/states"
	"github.com/direktiv/direktiv/pkg/model"
)

//
// README
//
// Here are the state logic implementations. If you're editing them or writing
// your own there are some things you should know.
//
// General Rules:
//
//   1. Under no circumstances should any functions here panic in production.
//	Panics here are not caught by the caller and will bring down the
//	server.
//
//   2. In all functions provided context.Context objects as an argument the
//	implementation must identify areas of logic that could run for a long
//	time and ensure that the logic can break out promptly if the context
// 	expires.

type stateChild struct {
	ID          string
	Type        string
	ServiceName string
}

type stateLogic interface {
	GetID() string
	GetType() model.StateType
	GetLog() interface{}
	GetMetadata() interface{}
	ErrorDefinitions() []model.ErrorDefinition

	Deadline(ctx context.Context) time.Time
	Run(ctx context.Context, wakedata []byte) (transition *states.Transition, err error)
}
