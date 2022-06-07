package flow

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/direktiv/direktiv/pkg/model"
)

type eventsXorStateLogic struct {
	*model.EventsXorState
	workflow *model.Workflow
}

func initEventsXorStateLogic(wf *model.Workflow, state model.State) (stateLogic, error) {

	eventsXor, ok := state.(*model.EventsXorState)
	if !ok {
		return nil, NewInternalError(errors.New("bad state object"))
	}

	sl := new(eventsXorStateLogic)
	sl.EventsXorState = eventsXor
	sl.workflow = wf

	return sl, nil

}

func (sl *eventsXorStateLogic) Deadline(ctx context.Context, engine *engine, im *instanceMemory) time.Time {
	return deadlineFromString(ctx, engine, im, sl.Timeout)
}

func (sl *eventsXorStateLogic) LivingChildren(ctx context.Context, engine *engine, im *instanceMemory) []stateChild {
	return nil
}

func (sl *eventsXorStateLogic) listenForEvents(ctx context.Context, engine *engine, im *instanceMemory) error {

	if im.GetMemory() != nil {
		return NewInternalError(errors.New("got unexpected savedata"))
	}

	var events []*model.ConsumeEventDefinition
	for i := range sl.Events {

		var err error
		event := new(model.ConsumeEventDefinition)
		event.Type = sl.Events[i].Event.Type
		event.Context = make(map[string]interface{})
		for k, v := range sl.Events[i].Event.Context {
			var x interface{}
			x, err = jqOne(im.data, v)
			if err != nil {
				err = NewUncatchableError("direktiv.event.jq", "failed to process event context key '%s': %v", k, err)
				return err
			}
			event.Context[k] = x
		}

		events = append(events, event)

	}

	err := engine.events.deleteInstanceEventListeners(ctx, im.in)
	if err != nil {
		return err
	}

	err = engine.events.listenForEvents(ctx, im, events, false)
	if err != nil {
		return err
	}

	return nil

}

func (sl *eventsXorStateLogic) Run(ctx context.Context, engine *engine, im *instanceMemory, wakedata []byte) (transition *stateTransition, err error) {

	if len(wakedata) == 0 {
		err = sl.listenForEvents(ctx, engine, im)
		return
	}

	events := make([]*cloudevents.Event, 0)
	err = json.Unmarshal(wakedata, &events)
	if err != nil {
		return
	}

	if len(events) != 1 {
		err = NewInternalError(errors.New("incorrect number of events returned"))
		return
	}

	for _, event := range events {

		err = im.StoreData(event.Type(), event)
		if err != nil {
			return
		}

		for i := 0; i < len(sl.Events); i++ {
			if sl.Events[i].Event.Type == event.Type() {
				transition = &stateTransition{
					Transform: sl.Events[i].Transform,
					NextState: sl.Events[i].Transition,
				}
				break
			}
		}

	}

	if transition == nil {
		err = NewInternalError(errors.New("got the wrong type of event back"))
		return
	}

	return

}
