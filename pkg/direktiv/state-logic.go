package direktiv

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/vorteil/direktiv/pkg/model"
)

const maxParallelActions = 10

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
	Transform string
}

type stateChild struct {
	Id   string
	Type string
}

type stateLogic interface {
	ID() string
	Type() string
	Deadline() time.Time
	ErrorCatchers() []model.ErrorDefinition
	Run(ctx context.Context, instance *workflowLogicInstance, savedata, wakedata []byte) (transition *stateTransition, err error)
	LivingChildren(savedata []byte) []stateChild
}

// -------------- Noop State --------------

type noopStateLogic struct {
	state *model.NoopState
}

func initNoopStateLogic(wf *model.Workflow, state model.State) (stateLogic, error) {

	noop, ok := state.(*model.NoopState)
	if !ok {
		return nil, NewInternalError(errors.New("bad state object"))
	}

	sl := new(noopStateLogic)
	sl.state = noop

	return sl, nil

}

func (sl *noopStateLogic) Type() string {
	return model.StateTypeNoop.String()
}

func (sl *noopStateLogic) Deadline() time.Time {
	return time.Now().Add(time.Second * 5)
}

func (sl *noopStateLogic) ErrorCatchers() []model.ErrorDefinition {
	return sl.state.ErrorDefinitions()
}

func (sl *noopStateLogic) ID() string {
	return sl.state.ID
}

func (sl *noopStateLogic) LivingChildren(savedata []byte) []stateChild {
	return nil
}

func (sl *noopStateLogic) Run(ctx context.Context, instance *workflowLogicInstance, savedata, wakedata []byte) (transition *stateTransition, err error) {

	if len(savedata) != 0 {
		err = NewInternalError(errors.New("got unexpected savedata"))
		return
	}

	if len(wakedata) != 0 {
		err = NewInternalError(errors.New("got unexpected wakedata"))
		return
	}

	object, err := jqObject(instance.data, ".")
	if err != nil {
		return
	}

	data, err := json.MarshalIndent(object, "", "  ")
	if err != nil {
		err = NewInternalError(fmt.Errorf("failed to marshal state data: %w", err))
		return
	}

	instance.Log("State data:\n%s", data)

	transition = &stateTransition{
		Transform: sl.state.Transform,
		NextState: sl.state.Transition,
	}

	return

}

type multiactionTuple struct {
	ID       string
	Complete bool
	Type     string
	Results  interface{}
}

func extractEventPayload(event *cloudevents.Event) (interface{}, error) {

	var x interface{}
	var err error

	if event.DataContentType() == "application/json" || event.DataContentType() == "" {
		err = json.Unmarshal(event.Data(), &x)
		if err != nil {
			return nil, NewInternalError(fmt.Errorf("Invalid json payload for event: %v", err))
		}
	} else {
		x = base64.StdEncoding.EncodeToString(event.Data())
	}

	return x, nil

}
