package states

import (
	"context"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/pkg/model"
)

var stateInitializers map[model.StateType]func(instance Instance, state model.State) (Logic, error)

func RegisterState(st model.StateType, initializer func(instance Instance, state model.State) (Logic, error)) {
	if stateInitializers == nil {
		stateInitializers = make(map[model.StateType]func(instance Instance, state model.State) (Logic, error))
	}

	if _, exists := stateInitializers[st]; exists {
		panic(fmt.Errorf("attempted to register duplicate state initializer"))
	}

	stateInitializers[st] = initializer
}

func StateLogic(instance Instance, state model.State) (Logic, error) {
	init, exists := stateInitializers[state.GetType()]
	if !exists {
		return nil, fmt.Errorf("cannot resolve state type: %s", state.GetType().String())
	}

	logic, err := init(instance, state)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize state logic: %w", err)
	}

	return logic, nil
}

type Logic interface {
	GetID() string
	GetType() model.StateType
	GetLog() interface{}
	GetMetadata() interface{}
	ErrorDefinitions() []model.ErrorDefinition
	GetMemory() interface{}
	Deadline(ctx context.Context) time.Time
	Run(ctx context.Context, wakedata []byte) (*Transition, error)
	LivingChildren(ctx context.Context) []*ChildInfo
}

type Transition struct {
	NextState string
	Transform interface{}
}
