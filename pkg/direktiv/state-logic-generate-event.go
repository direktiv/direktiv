package direktiv

import (
	"context"
	"encoding/base64"
	"errors"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
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

	event.SetType(sl.state.Event.Type)
	event.SetSource(sl.state.Event.Source)

	var x interface{}
	x, err = jqOne(instance.data, sl.state.Event.Data)
	if err != nil {
		return
	}

	var data []byte

	ctype := sl.state.Event.DataContentType
	if s, ok := x.(string); ok && ctype != "" && ctype != "application/json" || ctype == "" {
		data, err = base64.StdEncoding.DecodeString(s)
		if err != nil {
			instance.Log("Unable to decode results as a base64 encoded string. Reverting to JSON.")
		}
		event.SetData(ctype, data)
	}

	if data == nil {
		event.SetData("application/json", x)
	}

	// TODO: sl.state.Event.Context
	instance.Log("Context information not generated.")

	data, err = event.MarshalJSON()
	if err != nil {
		return
	}

	instance.Log("Broadcasting event: %s.", event.ID())

	_, err = instance.engine.grpcIngress.BroadcastEvent(ctx, &ingress.BroadcastEventRequest{
		Namespace:  &instance.namespace,
		Cloudevent: data,
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
