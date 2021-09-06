package direktiv

import (
	"context"
	"encoding/base64"
	"errors"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/segmentio/ksuid"
	"github.com/senseyeio/duration"
	"github.com/vorteil/direktiv/pkg/ingress"
	"github.com/vorteil/direktiv/pkg/model"
)

// -------------- GenerateEvent State --------------

type generateEventStateLogic struct {
	state *model.GenerateEventState
}

func initGenerateEventStateLogic(wf *model.Workflow, state model.State) (stateLogic, error) {

	gevent, ok := state.(*model.GenerateEventState)
	if !ok {
		return nil, NewInternalError(errors.New("bad state object"))
	}

	sl := new(generateEventStateLogic)
	sl.state = gevent

	return sl, nil

}

func (sl *generateEventStateLogic) Type() string {
	return model.StateTypeGenerateEvent.String()
}

func (sl *generateEventStateLogic) Deadline() time.Time {
	return time.Now().Add(time.Second * 5)
}

func (sl *generateEventStateLogic) ErrorCatchers() []model.ErrorDefinition {
	return sl.state.ErrorDefinitions()
}

func (sl *generateEventStateLogic) ID() string {
	return sl.state.ID
}

func (sl *generateEventStateLogic) LivingChildren(savedata []byte) []stateChild {
	return nil
}

func (sl *generateEventStateLogic) LogJQ() interface{} {
	return sl.state.Log
}

func (sl *generateEventStateLogic) Run(ctx context.Context, instance *workflowLogicInstance, savedata, wakedata []byte) (transition *stateTransition, err error) {

	if len(savedata) != 0 {
		err = NewInternalError(errors.New("got unexpected savedata"))
		return
	}

	if len(wakedata) != 0 {
		err = NewInternalError(errors.New("got unexpected wakedata"))
		return
	}

	event := cloudevents.NewEvent(cloudevents.VersionV03)

	uid := ksuid.New()
	event.SetID(uid.String())
	event.SetType(sl.state.Event.Type)
	event.SetSource(sl.state.Event.Source)

	var x interface{}
	x, err = jqOne(instance.data, sl.state.Event.Data)
	if err != nil {
		return
	}

	var data []byte

	ctype := sl.state.Event.DataContentType
	if s, ok := x.(string); ok && ctype != "" && ctype != "application/json" {
		data, err = base64.StdEncoding.DecodeString(s)
		if err != nil {
			instance.Log(ctx, "Unable to decode results as a base64 encoded string. Reverting to JSON.")
		}
		err = event.SetData(ctype, data)
		if err != nil {
			instance.Log(ctx, "Unable to set event data: %v", err)
		}
	}

	if data == nil {
		err = event.SetData("application/json", x)
		if err != nil {
			instance.Log(ctx, "Unable to set event data: %v", err)
		}
	}

	for k, v := range sl.state.Event.Context {
		var x interface{}
		x, err = jqOne(instance.data, v)
		if err != nil {
			err = NewUncatchableError("direktiv.event.jq", "failed to process event context key '%s': %v", k, err)
			return
		}

		instance.Log(ctx, "Adding context %v: %v", k, x)
		err = event.Context.SetExtension(k, x)
		if err != nil {
			instance.Log(ctx, "Unable to set event extension: %v", err)
		}
	}

	data, err = event.MarshalJSON()
	if err != nil {
		err = NewUncatchableError("direktiv.event.jq", "failed to marshal event: %v", err)
		return
	}
	instance.Log(ctx, "Broadcasting event: %s.", event.ID())

	var dd int64

	if len(sl.state.Delay) == 0 {
		dd = 60
		instance.eventQueue = append(instance.eventQueue, event.ID())
	} else if sl.state.Delay != "immediate" {
		d, _ := duration.ParseISO8601(sl.state.Delay)
		t := d.Shift(time.Unix(0, 0).UTC())
		dd = t.Unix()
	}

	appLog.Debugf("event fires in %d seconds", dd)

	_, err = instance.engine.ingressClient.BroadcastEvent(ctx, &ingress.BroadcastEventRequest{
		Namespace:  &instance.namespace,
		Cloudevent: data,
		Timer:      &dd,
	})
	if err != nil {
		return
	}

	transition = &stateTransition{
		Transform: sl.state.Transform,
		NextState: sl.state.Transition,
	}

	return

}
