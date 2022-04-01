package flow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/direktiv/direktiv/pkg/model"
)

type eventsAndStateLogic struct {
	state    *model.EventsAndState
	workflow *model.Workflow
}

func initEventsAndStateLogic(wf *model.Workflow, state model.State) (stateLogic, error) {

	eventsAnd, ok := state.(*model.EventsAndState)
	if !ok {
		return nil, NewInternalError(errors.New("bad state object"))
	}
	sl := new(eventsAndStateLogic)
	sl.state = eventsAnd
	sl.workflow = wf

	return sl, nil

}

func (sl *eventsAndStateLogic) Type() string {
	return model.StateTypeEventsAnd.String()
}

func (sl *eventsAndStateLogic) Deadline(ctx context.Context, engine *engine, im *instanceMemory) time.Time {
	return deadlineFromString(ctx, engine, im, sl.state.Timeout)
}

func (sl *eventsAndStateLogic) ErrorCatchers() []model.ErrorDefinition {
	return sl.state.ErrorDefinitions()
}

func (sl *eventsAndStateLogic) ID() string {
	return sl.state.ID
}

func (sl *eventsAndStateLogic) LivingChildren(ctx context.Context, engine *engine, im *instanceMemory) []stateChild {
	return nil
}

func (sl *eventsAndStateLogic) listenForEvents(ctx context.Context, engine *engine, im *instanceMemory) error {

	if im.GetMemory() != nil {
		return NewInternalError(errors.New("got unexpected savedata"))
	}

	var events []*model.ConsumeEventDefinition
	for i := range sl.state.Events {

		var err error
		event := new(model.ConsumeEventDefinition)
		event.Type = sl.state.Events[i].Type
		event.Context = make(map[string]interface{})
		for k, v := range sl.state.Events[i].Context {
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

	err = engine.events.listenForEvents(ctx, im, events, true)
	if err != nil {
		return err
	}

	return nil

}

func (sl *eventsAndStateLogic) LogJQ() interface{} {
	return sl.state.Log
}

func (sl *eventsAndStateLogic) MetadataJQ() interface{} {
	return sl.state.Metadata
}

func (sl *eventsAndStateLogic) Run(ctx context.Context, engine *engine, im *instanceMemory, wakedata []byte) (transition *stateTransition, err error) {

	if len(wakedata) == 0 {
		err = sl.listenForEvents(ctx, engine, im)
		return
	}

	events := make([]*cloudevents.Event, 0)

	err = json.Unmarshal(wakedata, &events)
	if err != nil {
		return
	}

	if len(events) != len(sl.state.Events) {
		err = NewInternalError(errors.New("incorrect number of events returned"))
		return
	}

	inMap := make(map[string]*cloudevents.Event)

	for a := range events {

		_, ok := inMap[events[a].Type()]
		k := events[a].Type()

		// duplicate, add a counter of the index on
		if ok {
			k = fmt.Sprintf("%s.%d", events[a].Type(), a)
		}
		inMap[k] = events[a]

	}

	// now stroe it
	for k, v := range inMap {
		err = im.StoreData(k, v)
		if err != nil {
			return
		}
	}

	transition = &stateTransition{
		Transform: sl.state.Transform,
		NextState: sl.state.Transition,
	}

	return

}
