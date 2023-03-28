package states

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	log "github.com/direktiv/direktiv/pkg/flow/internallogger"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/senseyeio/duration"
)

func init() {
	RegisterState(model.StateTypeEventsXor, EventsXor)
}

type eventsXorLogic struct {
	*model.EventsXorState
	Instance
}

func EventsXor(instance Instance, state model.State) (Logic, error) {
	eventsXor, ok := state.(*model.EventsXorState)
	if !ok {
		return nil, derrors.NewInternalError(errors.New("bad state object"))
	}

	sl := new(eventsXorLogic)
	sl.Instance = instance
	sl.EventsXorState = eventsXor

	return sl, nil
}

func (logic *eventsXorLogic) Deadline(ctx context.Context) time.Time {
	d, err := duration.ParseISO8601(logic.Timeout)
	if err != nil {
		logic.Log(ctx, log.Error, "failed to parse duration: %v", err)
		return time.Now().Add(DefaultLongDeadline)
	}

	t := d.Shift(time.Now().Add(DefaultShortDeadline))

	return t
}

func (logic *eventsXorLogic) Run(ctx context.Context, wakedata []byte) (*Transition, error) {
	first, err := scheduleTwice(logic, wakedata)
	if err != nil {
		return nil, err
	}

	if first {

		var events []*model.ConsumeEventDefinition

		for i := range logic.Events {

			event := new(model.ConsumeEventDefinition)
			event.Type = logic.Events[i].Event.Type
			event.Context = make(map[string]interface{})

			for k, v := range logic.Events[i].Event.Context {

				x, err := jqOne(logic.GetInstanceData(), v)
				if err != nil {
					return nil, derrors.NewUncatchableError("direktiv.event.jq", "failed to process event context key '%s': %v", k, err)
				}

				event.Context[k] = x

			}

			events = append(events, event)

		}

		err = logic.ListenForEvents(ctx, events, false)
		if err != nil {
			return nil, err
		}

		return nil, nil

	}

	events := make([]*cloudevents.Event, 0)

	err = json.Unmarshal(wakedata, &events)
	if err != nil {
		return nil, err
	}

	if len(events) != 1 {
		return nil, derrors.NewInternalError(errors.New("incorrect number of events returned"))
	}

	for _, event := range events {

		err = logic.StoreData(event.Type(), event)
		if err != nil {
			return nil, err
		}

		for i := 0; i < len(logic.Events); i++ {
			if logic.Events[i].Event.Type == event.Type() {
				return &Transition{
					Transform: logic.Events[i].Transform,
					NextState: logic.Events[i].Transition,
				}, nil
			}
		}

	}

	return nil, derrors.NewInternalError(errors.New("got the wrong type of event back"))
}
