package states

import (
	"context"
	"encoding/base64"
	"errors"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/google/uuid"
	"github.com/senseyeio/duration"
)

func init() {
	RegisterState(model.StateTypeGenerateEvent, GenerateEvent)
}

type generateEventLogic struct {
	*model.GenerateEventState
	Instance
}

func GenerateEvent(instance Instance, state model.State) (Logic, error) {
	generateEvent, ok := state.(*model.GenerateEventState)
	if !ok {
		return nil, derrors.NewInternalError(errors.New("bad state object"))
	}

	sl := new(generateEventLogic)
	sl.Instance = instance
	sl.GenerateEventState = generateEvent

	return sl, nil
}

func (logic *generateEventLogic) Run(ctx context.Context, wakedata []byte) (*Transition, error) {
	err := scheduleOnce(logic, wakedata)
	if err != nil {
		return nil, err
	}

	event := cloudevents.NewEvent(cloudevents.VersionV1)

	uid := uuid.New()
	event.SetID(uid.String())
	event.SetType(logic.Event.Type)
	event.SetSource(logic.Event.Source)

	var x interface{}
	x, err = jqOne(logic.GetInstanceData(), logic.Event.Data)
	if err != nil {
		return nil, err
	}

	var data []byte

	ctype := logic.Event.DataContentType
	if s, ok := x.(string); ok && ctype != "" && ctype != "application/json" {
		data, err = base64.StdEncoding.DecodeString(s)

		// trying to decode from base64, if it fails use it "as-is", e.g. plain-text
		if err != nil {
			err = event.SetData(ctype, s)
		} else {
			err = event.SetData(ctype, data)
		}
		if err != nil {
			// logic.Log(ctx, log.Error, "Unable to set event data: %v", err)
		}
	}

	if data == nil {
		err = event.SetData("application/json", x)
		if err != nil {
			// logic.Log(ctx, log.Error, "Unable to set event data: %v", err)
		}
	}

	for k, v := range logic.Event.Context {
		x, err := jqOne(logic.GetInstanceData(), v)
		if err != nil {
			return nil, derrors.NewUncatchableError("direktiv.event.jq", "failed to process event context key '%s': %v", k, err)
		}

		// logic.Log(ctx, log.Debug, "Adding context %v: %v", k, x)

		err = event.Context.SetExtension(k, x)
		if err != nil {
			// logic.Log(ctx, log.Error, "Unable to set event extension: %v", err)
		}
	}

	// logic.Log(ctx, log.Info, "Broadcasting event type:%s/source:%s to this namespace.", event.Type(), event.Source())

	var dd int64

	if len(logic.Delay) != 0 && logic.Delay != "immediate" {
		d, _ := duration.ParseISO8601(logic.Delay)
		t := d.Shift(time.Now().UTC())
		dd = t.Unix()
	}

	err = logic.BroadcastCloudevent(ctx, &event, dd)
	if err != nil {
		return nil, err
	}

	return &Transition{
		Transform: logic.Transform,
		NextState: logic.Transition,
	}, nil
}
