package flow

import (
	"context"
	"encoding/base64"
	"errors"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/google/uuid"
	"github.com/senseyeio/duration"
)

// -------------- GenerateEvent State --------------

type generateEventStateLogic struct {
	*model.GenerateEventState
}

func initGenerateEventStateLogic(wf *model.Workflow, state model.State) (stateLogic, error) {

	gevent, ok := state.(*model.GenerateEventState)
	if !ok {
		return nil, NewInternalError(errors.New("bad state object"))
	}

	sl := new(generateEventStateLogic)
	sl.GenerateEventState = gevent

	return sl, nil

}

func (sl *generateEventStateLogic) Deadline(ctx context.Context, engine *engine, im *instanceMemory) time.Time {
	return time.Now().Add(time.Second * 5)
}

func (sl *generateEventStateLogic) LivingChildren(ctx context.Context, engine *engine, im *instanceMemory) []stateChild {
	return nil
}

func (sl *generateEventStateLogic) Run(ctx context.Context, engine *engine, im *instanceMemory, wakedata []byte) (transition *stateTransition, err error) {

	if im.GetMemory() != nil {
		err = NewInternalError(errors.New("got unexpected savedata"))
		return
	}

	if len(wakedata) != 0 {
		err = NewInternalError(errors.New("got unexpected wakedata"))
		return
	}

	event := cloudevents.NewEvent(cloudevents.VersionV1)

	uid := uuid.New()
	event.SetID(uid.String())
	event.SetType(sl.Event.Type)
	event.SetSource(sl.Event.Source)

	var x interface{}
	x, err = jqOne(im.data, sl.Event.Data)
	if err != nil {
		return
	}

	var data []byte

	ctype := sl.Event.DataContentType
	if s, ok := x.(string); ok && ctype != "" && ctype != "application/json" {
		data, err = base64.StdEncoding.DecodeString(s)

		// trying to decode from base64, if it fails use it "as-is", e.g. plain-text
		if err != nil {
			err = event.SetData(ctype, s)
		} else {
			err = event.SetData(ctype, data)
		}

		if err != nil {
			engine.logToInstance(ctx, time.Now(), im.in, "Unable to set event data: %v", err)
		}
	}

	if data == nil {
		err = event.SetData("application/json", x)
		if err != nil {
			engine.logToInstance(ctx, time.Now(), im.in, "Unable to set event data: %v", err)
		}
	}

	for k, v := range sl.Event.Context {
		var x interface{}
		x, err = jqOne(im.data, v)
		if err != nil {
			err = NewUncatchableError("direktiv.event.jq", "failed to process event context key '%s': %v", k, err)
			return
		}
		// event.Context[k] = x
		engine.logToInstance(ctx, time.Now(), im.in, "Adding context %v: %v", k, x)
		err = event.Context.SetExtension(k, x)
		if err != nil {
			engine.logToInstance(ctx, time.Now(), im.in, "Unable to set event extension: %v", err)
		}
	}

	engine.logToInstance(ctx, time.Now(), im.in, "Broadcasting event: %s.", event.ID())

	var dd int64

	if len(sl.Delay) == 0 {
		dd = 60
		im.eventQueue = append(im.eventQueue, event.ID())
	} else if sl.Delay != "immediate" {
		d, _ := duration.ParseISO8601(sl.Delay)
		t := d.Shift(time.Unix(0, 0).UTC())
		dd = t.Unix()
	}

	// engine.sugar.Debugf("event fires in %d seconds", dd)

	err = engine.events.BroadcastCloudevent(ctx, im.in.Edges.Namespace, &event, dd)
	if err != nil {
		return
	}

	transition = &stateTransition{
		Transform: sl.Transform,
		NextState: sl.Transition,
	}

	return

}
