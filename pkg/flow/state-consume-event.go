package flow

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/direktiv/direktiv/pkg/model"
)

type consumeEventStateLogic struct {
	*model.ConsumeEventState
	workflow *model.Workflow
}

func initConsumeEventStateLogic(wf *model.Workflow, state model.State) (stateLogic, error) {

	cevent, ok := state.(*model.ConsumeEventState)
	if !ok {
		return nil, NewInternalError(errors.New("bad state object"))
	}

	sl := new(consumeEventStateLogic)
	sl.ConsumeEventState = cevent
	sl.workflow = wf

	return sl, nil

}

func (sl *consumeEventStateLogic) Deadline(ctx context.Context, engine *engine, im *instanceMemory) time.Time {
	return deadlineFromString(ctx, engine, im, sl.Timeout)
}

func (sl *consumeEventStateLogic) LivingChildren(ctx context.Context, engine *engine, im *instanceMemory) []stateChild {
	return nil
}

func (sl *consumeEventStateLogic) Run(ctx context.Context, engine *engine, im *instanceMemory, wakedata []byte) (transition *stateTransition, err error) {

	if len(wakedata) == 0 {

		// first part

		if im.GetMemory() != nil {
			err = NewInternalError(errors.New("got unexpected savedata"))
			return
		}

		var events []*model.ConsumeEventDefinition

		event := new(model.ConsumeEventDefinition)
		event.Type = sl.Event.Type
		event.Context = make(map[string]interface{})
		for k, v := range sl.Event.Context {
			var x interface{}
			x, err = jqOne(im.data, v)
			if err != nil {
				err = NewUncatchableError("direktiv.event.jq", "failed to process event context key '%s': %v", k, err)
				return
			}
			event.Context[k] = x
		}

		events = append(events, event)

		err = engine.events.deleteInstanceEventListeners(ctx, im.in)
		if err != nil {
			return
		}

		err = engine.events.listenForEvents(ctx, im, events, false)
		if err != nil {
			return
		}

		return

	}

	// second part

	events := make([]*cloudevents.Event, 0)
	err = json.Unmarshal(wakedata, &events)
	if err != nil {
		return
	}

	if len(events) == 0 {
		err = NewInternalError(errors.New("missing event in wakeup data"))
		return
	}

	if len(events) > 1 {
		err = NewInternalError(errors.New("multiple events returned when we were expecting just one"))
		return
	}

	for _, event := range events {

		err = im.StoreData(event.Type(), event)
		if err != nil {
			return
		}

	}

	transition = &stateTransition{
		Transform: sl.Transform,
		NextState: sl.Transition,
	}

	return

}
