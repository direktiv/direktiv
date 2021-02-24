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

type consumeEventStateLogic struct {
	state    *model.ConsumeEventState
	workflow *model.Workflow
}

func initConsumeEventStateLogic(wf *model.Workflow, state model.State) (stateLogic, error) {

	cevent, ok := state.(*model.ConsumeEventState)
	if !ok {
		return nil, NewInternalError(errors.New("bad state object"))
	}

	sl := new(consumeEventStateLogic)
	sl.state = cevent
	sl.workflow = wf

	return sl, nil

}

func (sl *consumeEventStateLogic) Type() string {
	return model.StateTypeConsumeEvent.String()
}

func (sl *consumeEventStateLogic) Deadline() time.Time {

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

func (sl *consumeEventStateLogic) ErrorCatchers() []model.ErrorDefinition {
	return sl.state.ErrorDefinitions()
}

func (sl *consumeEventStateLogic) ID() string {
	return sl.state.ID
}

func (sl *consumeEventStateLogic) LivingChildren(savedata []byte) []stateChild {
	return nil
}

func (sl *consumeEventStateLogic) Run(ctx context.Context, instance *workflowLogicInstance, savedata, wakedata []byte) (transition *stateTransition, err error) {

	if len(wakedata) == 0 {

		// first part

		if len(savedata) != 0 {
			err = NewInternalError(errors.New("got unexpected savedata"))
			return
		}

		var events []*model.ConsumeEventDefinition
		events = append(events, sl.state.Event)

		err = instance.engine.listenForEvents(ctx, instance, events, false)
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

	if len(events) == 0 {
		err = NewInternalError(errors.New("missing event in wakeup data"))
		return
	}

	if len(events) > 1 {
		err = NewInternalError(errors.New("multiple events returned when we were expecting just one"))
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

	}

	transition = &stateTransition{
		Transform: sl.state.Transform,
		NextState: sl.state.Transition,
	}

	return

}
