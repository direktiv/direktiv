package direktiv

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/senseyeio/duration"
	log "github.com/sirupsen/logrus"
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

	var t time.Time
	var d time.Duration

	d = time.Minute * 15

	if sl.state.Timeout != "" {
		dur, err := duration.ParseISO8601(sl.state.Timeout)
		if err != nil {
			// NOTE: validation should prevent this from ever happening
			log.Errorf("Got an invalid ISO8601 timeout: %v", err)
		} else {
			now := time.Now()
			later := dur.Shift(now)
			d = later.Sub(now)
		}
	}

	t = time.Now()
	t.Add(d)
	t.Add(time.Second * 5)

	return t

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

func (sl *eventsXorStateLogic) Run(ctx context.Context, instance *workflowLogicInstance, savedata, wakedata []byte) (transition *stateTransition, err error) {

	if len(wakedata) == 0 {

		// first part

		if len(savedata) != 0 {
			err = NewInternalError(errors.New("got unexpected savedata"))
			return
		}

		var events []*model.ConsumeEventDefinition
		for _, event := range sl.state.Events {
			events = append(events, &event.Event)
		}

		err = instance.engine.listenForEvents(ctx, instance, events, true)
		if err != nil {
			return
		}

		return

	}

	// second part

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

		if event.DataContentType() == "application/json" || event.DataContentType() == "" {
			err = json.Unmarshal(event.Data(), &x)
			if err != nil {
				err = NewInternalError(fmt.Errorf("Invalid json payload for event: %v", err))
				return
			}
		} else {
			x = base64.StdEncoding.EncodeToString(event.Data())
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
