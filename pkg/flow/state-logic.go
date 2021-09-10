package flow

import (
	"context"
	"time"

	"github.com/vorteil/direktiv/pkg/model"
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

type stateTransition struct {
	NextState string
	Transform interface{}
}

type stateChild struct {
	Id   string
	Type string
}

type stateLogic interface {
	ID() string
	Type() string
	Deadline(ctx context.Context, engine *engine, im *instanceMemory) time.Time
	ErrorCatchers() []model.ErrorDefinition
	Run(ctx context.Context, engine *engine, im *instanceMemory, wakedata []byte) (transition *stateTransition, err error)
	LivingChildren(ctx context.Context, engine *engine, im *instanceMemory) []stateChild
	LogJQ() interface{}
}
