package states

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	log "github.com/direktiv/direktiv/pkg/flow/internallogger"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/senseyeio/duration"
)

//nolint:gochecknoinits
func init() {
	RegisterState(model.StateTypeEventsAnd, EventsAnd)
}

type eventsAndLogic struct {
	*model.EventsAndState
	Instance
}

func EventsAnd(instance Instance, state model.State) (Logic, error) {
	eventsAnd, ok := state.(*model.EventsAndState)
	if !ok {
		return nil, derrors.NewInternalError(errors.New("bad state object"))
	}

	sl := new(eventsAndLogic)
	sl.Instance = instance
	sl.EventsAndState = eventsAnd

	return sl, nil
}

func (logic *eventsAndLogic) Deadline(ctx context.Context) time.Time {
	d, err := duration.ParseISO8601(logic.Timeout)
	if err != nil {
		logic.Log(ctx, log.Error, "failed to parse duration: %v", err)
		return time.Now().UTC().Add(DefaultLongDeadline)
	}

	t := d.Shift(time.Now().UTC().Add(DefaultShortDeadline))

	return t
}

func (logic *eventsAndLogic) Run(ctx context.Context, wakedata []byte) (*Transition, error) {
	first, err := scheduleTwice(logic, wakedata)
	if err != nil {
		return nil, err
	}

	if first {
		var events []*model.ConsumeEventDefinition

		for i := range logic.Events {
			event := new(model.ConsumeEventDefinition)
			event.Type = logic.Events[i].Type
			event.Context = make(map[string]interface{})

			for k, v := range logic.Events[i].Context {
				x, err := jqOne(logic.GetInstanceData(), v) //nolint:contextcheck
				if err != nil {
					return nil, derrors.NewUncatchableError("direktiv.event.jq", "failed to process event context key '%s': %v", k, err)
				}

				event.Context[k] = x
			}

			events = append(events, event)
		}

		err = logic.ListenForEvents(ctx, events, true)
		if err != nil {
			return nil, err
		}

		//nolint:nilnil
		return nil, nil
	}

	events := make([]*cloudevents.Event, 0)

	err = json.Unmarshal(wakedata, &events)
	if err != nil {
		return nil, err
	}

	if len(events) != len(logic.Events) {
		return nil, derrors.NewInternalError(errors.New("incorrect number of events returned"))
	}

	inMap := make(map[string]*cloudevents.Event)

	for a := range events {
		_, ok := inMap[events[a].Type()]
		k := events[a].Type()

		if ok {
			k = fmt.Sprintf("%s.%d", events[a].Type(), a)
		}

		inMap[k] = events[a]
	}

	for k, v := range inMap {
		err = logic.StoreData(k, v)
		if err != nil {
			return nil, err
		}
	}

	return &Transition{
		Transform: logic.Transform,
		NextState: logic.Transition,
	}, nil
}
