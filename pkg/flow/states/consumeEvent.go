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

//nolint:gochecknoinits
func init() {
	RegisterState(model.StateTypeConsumeEvent, ConsumeEvent)
}

type consumeEventLogic struct {
	*model.ConsumeEventState
	Instance
}

func ConsumeEvent(instance Instance, state model.State) (Logic, error) {
	consumeEvent, ok := state.(*model.ConsumeEventState)
	if !ok {
		return nil, derrors.NewInternalError(errors.New("bad state object"))
	}

	sl := new(consumeEventLogic)
	sl.Instance = instance
	sl.ConsumeEventState = consumeEvent

	return sl, nil
}

func (logic *consumeEventLogic) Deadline(ctx context.Context) time.Time {
	d, err := duration.ParseISO8601(logic.Timeout)
	if err != nil {
		logic.Log(ctx, log.Error, "failed to parse duration: %v", err)
		return time.Now().UTC().Add(DefaultLongDeadline)
	}

	t := d.Shift(time.Now().UTC().Add(DefaultShortDeadline))

	return t
}

func (logic *consumeEventLogic) Run(ctx context.Context, wakedata []byte) (*Transition, error) {
	first, err := scheduleTwice(logic, wakedata)
	if err != nil {
		return nil, err
	}

	if first {
		var events []*model.ConsumeEventDefinition

		event := new(model.ConsumeEventDefinition)
		event.Type = logic.Event.Type
		event.Context = make(map[string]interface{})

		for k, v := range logic.Event.Context {
			x, err := jqOne(logic.GetInstanceData(), v) //nolint:contextcheck
			if err != nil {
				return nil, derrors.NewUncatchableError("direktiv.event.jq", "failed to process event context key '%s': %v", k, err)
			}

			event.Context[k] = x
		}

		events = append(events, event)

		err = logic.ListenForEvents(ctx, events, false)
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

	if len(events) == 0 {
		return nil, derrors.NewInternalError(errors.New("missing event in wakeup data"))
	}

	if len(events) > 1 {
		return nil, derrors.NewInternalError(errors.New("multiple events returned when we were expecting just one"))
	}

	for _, event := range events {
		err = logic.StoreData(event.Type(), event)
		if err != nil {
			return nil, err
		}
	}

	return &Transition{
		Transform: logic.Transform,
		NextState: logic.Transition,
	}, nil
}
