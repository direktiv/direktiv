package direktiv

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/vorteil/direktiv/pkg/model"
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

func (sl *eventsAndStateLogic) Deadline() time.Time {
	return deadlineFromString(sl.state.Timeout)
}

func (sl *eventsAndStateLogic) ErrorCatchers() []model.ErrorDefinition {
	return sl.state.ErrorDefinitions()
}

func (sl *eventsAndStateLogic) ID() string {
	return sl.state.ID
}

func (sl *eventsAndStateLogic) LivingChildren(savedata []byte) []stateChild {
	return nil
}

func (sl *eventsAndStateLogic) listenForEvents(ctx context.Context, instance *workflowLogicInstance, savedata []byte) error {

	if len(savedata) != 0 {
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
	err := instance.engine.listenForEvents(ctx, instance, events, true)
	if err != nil {
		return err
	}

	return nil

}

func (sl *eventsAndStateLogic) LogJQ() interface{} {
	return sl.state.Log
}

func (sl *eventsAndStateLogic) Run(ctx context.Context, instance *workflowLogicInstance, savedata, wakedata []byte) (transition *stateTransition, err error) {

	if len(wakedata) == 0 {
		err = sl.listenForEvents(ctx, instance, savedata)
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

	}

	transition = &stateTransition{
		Transform: sl.state.Transform,
		NextState: sl.state.Transition,
	}

	return

}
