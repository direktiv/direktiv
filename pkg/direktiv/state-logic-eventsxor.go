package direktiv

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/vorteil/direktiv/pkg/model"
)

type eventsXorStateLogic struct {
	state    *model.EventsXorState
	workflow *model.Workflow
}

func initEventsXorStateLogic(wf *model.Workflow, state model.State) (stateLogic, error) {

	eventsXor, ok := state.(*model.EventsXorState)
	if !ok {
		return nil, NewInternalError(errors.New("bad state object"))
	}

	sl := new(eventsXorStateLogic)
	sl.state = eventsXor
	sl.workflow = wf

	return sl, nil

}

func (sl *eventsXorStateLogic) Type() string {
	return model.StateTypeEventsXor.String()
}

func (sl *eventsXorStateLogic) Deadline() time.Time {
	return deadlineFromString(sl.state.Timeout)
}

func (sl *eventsXorStateLogic) ErrorCatchers() []model.ErrorDefinition {
	return sl.state.ErrorDefinitions()
}

func (sl *eventsXorStateLogic) ID() string {
	return sl.state.ID
}

func (sl *eventsXorStateLogic) LivingChildren(savedata []byte) []stateChild {
	return nil
}

func (sl *eventsXorStateLogic) listenForEvents(ctx context.Context, instance *workflowLogicInstance, savedata []byte) error {

	if len(savedata) != 0 {
		return NewInternalError(errors.New("got unexpected savedata"))
	}

	var events []*model.ConsumeEventDefinition
	for i := range sl.state.Events {

		var err error
		event := new(model.ConsumeEventDefinition)
		event.Type = sl.state.Events[i].Event.Type
		event.Context = make(map[string]interface{})
		for k, v := range sl.state.Events[i].Event.Context {
			var x interface{}
			x, err = jqOne(instance.data, v)
			if err != nil {
				err = NewUncatchableError("direktiv.event.jq", "failed to process event context key '%s': %v", k, err)
				return err
			}
			event.Context[k] = x
		}

		events = append(events, event)

	}

	instance.engine.clearEventListeners(instance.rec)
	err := instance.engine.listenForEvents(ctx, instance, events, false)
	if err != nil {
		return err
	}

	return nil

}

func (sl *eventsXorStateLogic) LogJQ() interface{} {
	return sl.state.Log
}

func (sl *eventsXorStateLogic) Run(ctx context.Context, instance *workflowLogicInstance, savedata, wakedata []byte) (transition *stateTransition, err error) {

	if len(wakedata) == 0 {
		err = sl.listenForEvents(ctx, instance, savedata)
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

		var x interface{}

		x, err = extractEventPayload(event)
		if err != nil {
			return
		}

		err = instance.StoreData(event.Type(), x)
		if err != nil {
			return
		}

		for i := 0; i < len(sl.state.Events); i++ {
			if sl.state.Events[i].Event.Type == event.Type() {
				transition = &stateTransition{
					Transform: sl.state.Events[i].Transform,
					NextState: sl.state.Events[i].Transition,
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
