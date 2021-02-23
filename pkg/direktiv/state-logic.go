package direktiv

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/vorteil/direktiv/pkg/ingress"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/segmentio/ksuid"
	"github.com/senseyeio/duration"
	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/pkg/model"
	"github.com/xeipuuv/gojsonschema"
)

const maxParallelActions = 10

//
// README
//
// Here are the state logic implementations. If you're editing them or writing
// your own there are some things you should know.
//
// General Rules:
//
//   1. Under no circumstances should any functions here panic in production.
//	Panics here are not caught by the caller and will bring down the
//	server.
//
//   2. In all functions provided context.Context objects as an argument the
//	implementation must identify areas of logic that could run for a long
//	time and ensure that the logic can break out promptly if the context
// 	expires.

type stateTransition struct {
	NextState string
	Transform string
}

type stateChild struct {
	Id   string
	Type string
}

type stateLogic interface {
	ID() string
	Type() string
	Deadline() time.Time
	ErrorCatchers() []model.ErrorDefinition
	Run(ctx context.Context, instance *workflowLogicInstance, savedata, wakedata []byte) (transition *stateTransition, err error)
	LivingChildren(savedata []byte) []stateChild
}

// -------------- Noop State --------------

type noopStateLogic struct {
	state *model.NoopState
}

func initNoopStateLogic(wf *model.Workflow, state model.State) (stateLogic, error) {

	noop, ok := state.(*model.NoopState)
	if !ok {
		return nil, NewInternalError(errors.New("bad state object"))
	}

	sl := new(noopStateLogic)
	sl.state = noop

	return sl, nil

}

func (sl *noopStateLogic) Type() string {
	return model.StateTypeNoop.String()
}

func (sl *noopStateLogic) Deadline() time.Time {
	return time.Now().Add(time.Second * 5)
}

func (sl *noopStateLogic) ErrorCatchers() []model.ErrorDefinition {
	return sl.state.ErrorDefinitions()
}

func (sl *noopStateLogic) ID() string {
	return sl.state.ID
}

func (sl *noopStateLogic) LivingChildren(savedata []byte) []stateChild {
	return nil
}

func (sl *noopStateLogic) Run(ctx context.Context, instance *workflowLogicInstance, savedata, wakedata []byte) (transition *stateTransition, err error) {

	if len(savedata) != 0 {
		err = NewInternalError(errors.New("got unexpected savedata"))
		return
	}

	if len(wakedata) != 0 {
		err = NewInternalError(errors.New("got unexpected wakedata"))
		return
	}

	object, err := instance.JQObject(".")
	if err != nil {
		return
	}

	data, err := json.MarshalIndent(object, "", "  ")
	if err != nil {
		err = NewInternalError(fmt.Errorf("failed to marshal state data: %w", err))
		return
	}

	instance.Log("State data:\n%s", data)

	transition = &stateTransition{
		Transform: sl.state.Transform,
		NextState: sl.state.Transition,
	}

	return

}

// -------------- Action State --------------

type actionStateLogic struct {
	state    *model.ActionState
	workflow *model.Workflow
}

func initActionStateLogic(wf *model.Workflow, state model.State) (stateLogic, error) {

	action, ok := state.(*model.ActionState)
	if !ok {
		return nil, NewInternalError(errors.New("bad state object"))
	}

	sl := new(actionStateLogic)
	sl.state = action
	sl.workflow = wf

	return sl, nil

}

func (sl *actionStateLogic) Type() string {
	return model.StateTypeAction.String()
}

func (sl *actionStateLogic) Deadline() time.Time {

	if sl.state.Async {
		return time.Now().Add(time.Second * 5)
	}

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

func (sl *actionStateLogic) ErrorCatchers() []model.ErrorDefinition {
	return sl.state.ErrorDefinitions()
}

func (sl *actionStateLogic) ID() string {
	return sl.state.ID
}

func (sl *actionStateLogic) LivingChildren(savedata []byte) []stateChild {

	var err error
	var children = make([]stateChild, 0)

	if sl.state.Action.Function != "" {

		var uid ksuid.KSUID
		uid, err = ksuid.FromBytes(savedata)
		if err != nil {
			log.Error(err)
			return children
		}

		children = append(children, stateChild{
			Id:   uid.String(),
			Type: "isolate",
		})

	} else {

		id := string(savedata)

		children = append(children, stateChild{
			Id:   id,
			Type: "subflow",
		})

	}

	return children

}

func (sl *actionStateLogic) Run(ctx context.Context, instance *workflowLogicInstance, savedata, wakedata []byte) (transition *stateTransition, err error) {

	if len(wakedata) == 0 {

		// first part

		if len(savedata) != 0 {
			err = NewInternalError(errors.New("got unexpected savedata"))
			return
		}

		var input interface{}

		input, err = instance.JQ(".")
		if err != nil {
			return
		}

		m, ok := input.(map[string]interface{})
		if !ok {
			err = NewInternalError(errors.New("invalid state data"))
			return
		}

		if len(sl.state.Action.Secrets) > 0 {
			instance.Log("Decrypting secrets.")

			s := make(map[string]string)

			for _, name := range sl.state.Action.Secrets {

				var dd []byte
				dd, err = decryptedDataForNS(ctx, instance, instance.namespace, name)
				if err != nil {
					return
				}
				s[name] = string(dd)

			}

			m["secrets"] = s
		}

		input, err = jqPreferObject(m, sl.state.Action.Input)
		if err != nil {
			return
		}

		var inputData []byte

		inputData, err = json.Marshal(input)
		if err != nil {
			err = NewInternalError(err)
			return
		}

		if sl.state.Action.Function != "" {

			// container

			uid := ksuid.New()
			err = instance.Save(ctx, uid.Bytes())
			if err != nil {
				return
			}

			var fn *model.FunctionDefinition

			fn, err = sl.workflow.GetFunction(sl.state.Action.Function)
			if err != nil {
				err = NewInternalError(err)
				return
			}

			ar := new(actionRequest)
			ar.ActionID = uid.String()
			ar.Workflow.InstanceID = instance.id
			ar.Workflow.Namespace = instance.namespace
			ar.Workflow.State = sl.state.GetID()
			ar.Workflow.Step = instance.step
			ar.Container.Image = fn.Image
			ar.Container.Cmd = fn.Cmd
			ar.Container.Size = int32(fn.Size)

			// TODO: timeout
			ar.Container.Data = inputData
			ar.Container.Registries = make(map[string]string)

			// get registries
			ar.Container.Registries, err = getRegistries(instance.engine.server.config, instance.namespace)
			if err != nil {
				return
			}

			if sl.state.Async {

				instance.Log("Running function '%s' in fire-and-forget mode (async).", fn.ID)

				go func(ctx context.Context, instance *workflowLogicInstance, ar *actionRequest) {

					ar.Workflow.InstanceID = ""
					ar.Workflow.Namespace = ""
					ar.Workflow.State = ""
					ar.Workflow.Step = 0

					// get registries
					ar.Container.Registries, err = getRegistries(instance.engine.server.config, instance.namespace)
					if err != nil {
						return
					}

					err = instance.engine.doActionRequest(ctx, ar)
					if err != nil {
						return
					}

				}(ctx, instance, ar)

				transition = &stateTransition{
					Transform: sl.state.Transform,
					NextState: sl.state.Transition,
				}

				return

			} else {

				instance.Log("Sleeping until function '%s' returns.", fn.ID)

				err = instance.engine.doActionRequest(ctx, ar)
				if err != nil {
					return
				}

			}

		} else {

			// subflow

			caller := new(subflowCaller)
			caller.InstanceID = instance.id
			caller.State = sl.state.GetID()
			caller.Step = instance.step

			var subflowID string

			if sl.state.Async {

				subflowID, err = instance.engine.subflowInvoke(caller, instance.rec.InvokedBy, instance.namespace, sl.state.Action.Workflow, inputData)
				if err != nil {
					return
				}

				instance.Log("Running subflow '%s' in fire-and-forget mode (async).", subflowID)

				transition = &stateTransition{
					Transform: sl.state.Transform,
					NextState: sl.state.Transition,
				}

				return

			} else {

				subflowID, err = instance.engine.subflowInvoke(caller, instance.rec.InvokedBy, instance.namespace, sl.state.Action.Workflow, inputData)
				if err != nil {
					return
				}

				instance.Log("Sleeping until subflow '%s' returns.", subflowID)

				err = instance.Save(ctx, []byte(subflowID))
				if err != nil {
					return
				}

			}

		}

		return

	}

	// second part

	results := new(actionResultPayload)
	err = json.Unmarshal(wakedata, results)
	if err != nil {
		err = NewInternalError(err)
		return
	}

	if sl.state.Action.Function != "" {

		var uid ksuid.KSUID
		uid, err = ksuid.FromBytes(savedata)
		if err != nil {
			err = NewInternalError(err)
			return
		}

		if results.ActionID != uid.String() {
			err = NewInternalError(errors.New("incorrect action ID"))
			return
		}

		instance.Log("Function '%s' returned.", sl.state.Action.Function)

	} else {

		id := string(savedata)
		if results.ActionID != id {
			err = NewInternalError(errors.New("incorrect subflow action ID"))
			return
		}

		instance.Log("Subflow '%s' returned.", id)

	}

	if results.ErrorCode != "" {

		instance.Log("Action raised catchable error '%s': %s.", results.ErrorCode, results.ErrorMessage)

		err = NewCatchableError(results.ErrorCode, results.ErrorMessage)
		return
	}

	if results.ErrorMessage != "" {

		instance.Log("Action crashed due to an internal error.")

		err = NewInternalError(errors.New(results.ErrorMessage))
		return
	}

	var x interface{}
	err = json.Unmarshal(results.Output, &x)
	if err != nil {
		x = base64.StdEncoding.EncodeToString(results.Output)
	}

	err = instance.StoreData("return", x)
	if err != nil {
		err = NewInternalError(err)
		return
	}

	transition = &stateTransition{
		Transform: sl.state.Transform,
		NextState: sl.state.Transition,
	}

	return

}

// -------------- ConsumeEvent State --------------

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

// -------------- Delay State --------------

type delayStateLogic struct {
	state *model.DelayState
}

func initDelayStateLogic(wf *model.Workflow, state model.State) (stateLogic, error) {

	delay, ok := state.(*model.DelayState)
	if !ok {
		return nil, NewInternalError(errors.New("bad state object"))
	}

	sl := new(delayStateLogic)
	sl.state = delay
	return sl, nil

}

func (sl *delayStateLogic) Type() string {
	return model.StateTypeDelay.String()
}

func (sl *delayStateLogic) Deadline() time.Time {

	d, err := duration.ParseISO8601(sl.state.Duration)
	if err != nil {
		log.Errorf("failed to parse duration: %v", err)
		return time.Now()
	}

	t := d.Shift(time.Now().Add(time.Second * 5))
	return t

}

func (sl *delayStateLogic) ErrorCatchers() []model.ErrorDefinition {
	return sl.state.ErrorDefinitions()
}

func (sl *delayStateLogic) ID() string {
	return sl.state.GetID()
}

func (sl *delayStateLogic) LivingChildren(savedata []byte) []stateChild {
	return nil
}

func (sl *delayStateLogic) Run(ctx context.Context, instance *workflowLogicInstance, savedata, wakedata []byte) (transition *stateTransition, err error) {

	if len(savedata) != 0 {
		err = NewInternalError(errors.New("got unexpected savedata"))
		return
	}

	if len(wakedata) == 0 {

		var d duration.Duration
		d, err = duration.ParseISO8601(sl.state.Duration)
		if err != nil {
			err = NewInternalError(fmt.Errorf("failed to parse delay duration: %v", err))
			return
		}

		t := d.Shift(time.Now())

		err = instance.engine.sleep(instance.id, sl.ID(), instance.step, t)
		if err != nil {
			return
		}

		return

	} else if string(wakedata) == sleepWakedata {

		transition = &stateTransition{
			Transform: sl.state.Transform,
			NextState: sl.state.Transition,
		}

		return

	} else {

		err = NewInternalError(fmt.Errorf("unexpected wakedata for delay state: %s", wakedata))
		return

	}

}

// -------------- Error State --------------

type errorStateLogic struct {
	state *model.ErrorState
}

func initErrorStateLogic(wf *model.Workflow, state model.State) (stateLogic, error) {

	err, ok := state.(*model.ErrorState)
	if !ok {
		return nil, NewInternalError(errors.New("bad state object"))
	}

	sl := new(errorStateLogic)
	sl.state = err

	return sl, nil

}

func (sl *errorStateLogic) Type() string {
	return model.StateTypeError.String()
}

func (sl *errorStateLogic) Deadline() time.Time {
	return time.Now().Add(time.Second * 5)
}

func (sl *errorStateLogic) ErrorCatchers() []model.ErrorDefinition {
	return sl.state.ErrorDefinitions()
}

func (sl *errorStateLogic) ID() string {
	return sl.state.GetID()
}

func (sl *errorStateLogic) LivingChildren(savedata []byte) []stateChild {
	return nil
}

func (sl *errorStateLogic) Run(ctx context.Context, instance *workflowLogicInstance, savedata, wakedata []byte) (transition *stateTransition, err error) {

	a := make([]interface{}, len(sl.state.Args))

	for i := 0; i < len(a); i++ {
		var x interface{}
		x, err = instance.JQObject(sl.state.Args[i])
		if err != nil {
			return
		}
		a[i] = x
	}

	err = instance.Raise(ctx, NewCatchableError(sl.state.Error, sl.state.Message, a...))
	if err != nil {
		return
	}

	transition = &stateTransition{
		Transform: sl.state.Transform,
		NextState: sl.state.Transition,
	}

	return

}

// -------------- EventsAnd State --------------

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

func (sl *eventsAndStateLogic) ErrorCatchers() []model.ErrorDefinition {
	return sl.state.ErrorDefinitions()
}

func (sl *eventsAndStateLogic) ID() string {
	return sl.state.ID
}

func (sl *eventsAndStateLogic) LivingChildren(savedata []byte) []stateChild {
	return nil
}

func (sl *eventsAndStateLogic) Run(ctx context.Context, instance *workflowLogicInstance, savedata, wakedata []byte) (transition *stateTransition, err error) {

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

	if len(events) != len(sl.state.Events) {
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

	}

	transition = &stateTransition{
		Transform: sl.state.Transform,
		NextState: sl.state.Transition,
	}

	return

}

// -------------- EventsXor State --------------

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

// -------------- Foreach State --------------

type foreachStateLogic struct {
	state    *model.ForEachState
	workflow *model.Workflow
}

func initForEachStateLogic(wf *model.Workflow, state model.State) (stateLogic, error) {

	foreach, ok := state.(*model.ForEachState)
	if !ok {
		return nil, NewInternalError(errors.New("bad state object"))
	}

	sl := new(foreachStateLogic)
	sl.state = foreach
	sl.workflow = wf

	return sl, nil

}

func (sl *foreachStateLogic) Type() string {
	return model.StateTypeForEach.String()
}

func (sl *foreachStateLogic) Deadline() time.Time {

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

func (sl *foreachStateLogic) ErrorCatchers() []model.ErrorDefinition {
	return sl.state.ErrorDefinitions()
}

func (sl *foreachStateLogic) ID() string {
	return sl.state.ID
}

func (sl *foreachStateLogic) LivingChildren(savedata []byte) []stateChild {

	var err error
	var children = make([]stateChild, 0)

	var logics []multiactionTuple
	err = json.Unmarshal(savedata, &logics)
	if err != nil {
		log.Error(err)
		return children
	}

	for _, logic := range logics {
		if logic.Complete {
			continue
		}
		children = append(children, stateChild{
			Id:   logic.ID,
			Type: logic.Type,
		})
	}

	return children

}

func (sl *foreachStateLogic) Run(ctx context.Context, instance *workflowLogicInstance, savedata, wakedata []byte) (transition *stateTransition, err error) {

	if len(wakedata) == 0 {

		// first part

		logics := make([]multiactionTuple, 0)

		if len(savedata) != 0 {
			err = NewInternalError(errors.New("got unexpected savedata"))
			return
		}

		var x interface{}
		x, err = instance.JQ(sl.state.Array)
		if err != nil {
			return
		}

		var ok bool
		var array []interface{}
		if array, ok = x.([]interface{}); !ok {
			array = append(array, x)
		}

		instance.Log("Generated %d objects to loop over.", len(array))

		if len(array) > maxParallelActions {
			err = NewUncatchableError("direktiv.limits.parallel", "instance aborted for exceeding the maximum number of parallel actions (%d)", maxParallelActions)
			return
		}

		action := sl.state.Action

		for _, inputSource := range array {

			var input interface{}

			input, err = jqMustBeObject(inputSource, ".")
			if err != nil {
				return
			}

			m, ok := input.(map[string]interface{})
			if !ok {
				err = NewInternalError(errors.New("invalid state data"))
				return
			}

			if len(sl.state.Action.Secrets) > 0 {
				instance.Log("Decrypting secrets.")

				s := make(map[string]string)

				for _, name := range sl.state.Action.Secrets {
					var dd []byte
					dd, err = decryptedDataForNS(ctx, instance, instance.namespace, name)
					if err != nil {
						return
					}
					s[name] = string(dd)
				}

				m["secrets"] = s
			}

			input, err = jqPreferObject(m, action.Input)
			if err != nil {
				return
			}

			var inputData []byte

			inputData, err = json.Marshal(input)
			if err != nil {
				err = NewInternalError(err)
				return
			}

			if action.Function != "" {

				// container

				uid := ksuid.New()
				logics = append(logics, multiactionTuple{
					ID:   uid.String(),
					Type: "isolate",
				})

				var fn *model.FunctionDefinition

				fn, err = sl.workflow.GetFunction(action.Function)
				if err != nil {
					err = NewInternalError(err)
					return
				}

				ar := new(actionRequest)
				ar.ActionID = uid.String()
				ar.Workflow.InstanceID = instance.id
				ar.Workflow.Namespace = instance.namespace
				ar.Workflow.State = sl.state.GetID()
				ar.Workflow.Step = instance.step
				ar.Container.Image = fn.Image
				ar.Container.Cmd = fn.Cmd
				ar.Container.Size = int32(fn.Size)

				// TODO: timeout
				ar.Container.Data = inputData

				// get registries
				ar.Container.Registries, err = getRegistries(instance.engine.server.config, instance.namespace)
				if err != nil {
					return
				}

				err = instance.engine.doActionRequest(ctx, ar)
				if err != nil {
					return
				}

			} else {

				// subflow

				caller := new(subflowCaller)
				caller.InstanceID = instance.id
				caller.State = sl.state.GetID()
				caller.Step = instance.step

				var subflowID string

				// TODO: log subflow instance IDs

				subflowID, err = instance.engine.subflowInvoke(caller, instance.rec.InvokedBy, instance.namespace, action.Workflow, inputData)
				if err != nil {
					return
				}

				logics = append(logics, multiactionTuple{
					ID:   subflowID,
					Type: "subflow",
				})

			}

		}

		var data []byte
		data, err = json.Marshal(logics)
		if err != nil {
			err = NewInternalError(err)
			return
		}

		err = instance.Save(ctx, data)
		if err != nil {
			return
		}

		return

	}

	// second part

	results := new(actionResultPayload)
	err = json.Unmarshal(wakedata, results)
	if err != nil {
		err = NewInternalError(err)
		return
	}

	var logics []multiactionTuple
	err = json.Unmarshal(savedata, &logics)
	if err != nil {
		err = NewInternalError(err)
		return
	}

	var found bool
	var idx int
	var completed int

	for i, lid := range logics {

		if lid.ID == results.ActionID {
			found = true
			if lid.Complete {
				err = NewInternalError(fmt.Errorf("action '%s' already completed", lid.ID))
				return
			}
			logics[i].Complete = true
			lid.Complete = true
			idx = i
		}

		if lid.Complete {
			completed++
		}

	}

	if !found {
		err = NewInternalError(fmt.Errorf("action '%s' wasn't expected", results.ActionID))
		return
	}

	instance.Log("Action returned. (%d/%d)", completed, len(logics))

	if results.ErrorCode != "" {
		instance.Log("Action returned catchable error '%s': %s.", results.ErrorCode, results.ErrorMessage)
		err = NewCatchableError(results.ErrorCode, results.ErrorMessage)
		return
	}

	if results.ErrorMessage != "" {
		instance.Log("Action crashed due to an internal error.")
		err = NewInternalError(errors.New(results.ErrorMessage))
		return
	}

	var x interface{}
	err = json.Unmarshal(results.Output, &x)
	if err != nil {
		x = base64.StdEncoding.EncodeToString(results.Output)
	}

	logics[idx].Results = x

	var ready bool
	if completed == len(logics) {
		ready = true
	}

	if ready {

		var results []interface{}
		for i := range logics {
			results = append(results, logics[i].Results)
		}

		err = instance.StoreData("return", results)
		if err != nil {
			err = NewInternalError(err)
			return
		}

		transition = &stateTransition{
			Transform: sl.state.Transform,
			NextState: sl.state.Transition,
		}

		return

	}

	var data []byte
	data, err = json.Marshal(logics)
	if err != nil {
		err = NewInternalError(err)
		return
	}

	err = instance.Save(ctx, data)
	if err != nil {
		return
	}

	return

}

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
	x, err = instance.JQ(sl.state.Event.Data)
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

// -------------- Parallel State --------------

type multiactionTuple struct {
	ID       string
	Complete bool
	Type     string
	Results  interface{}
}

type parallelStateLogic struct {
	state    *model.ParallelState
	workflow *model.Workflow
}

func initParallelStateLogic(wf *model.Workflow, state model.State) (stateLogic, error) {

	parallel, ok := state.(*model.ParallelState)
	if !ok {
		return nil, NewInternalError(errors.New("bad state object"))
	}

	sl := new(parallelStateLogic)
	sl.state = parallel
	sl.workflow = wf

	return sl, nil

}

func (sl *parallelStateLogic) Type() string {
	return model.StateTypeParallel.String()
}

func (sl *parallelStateLogic) Deadline() time.Time {

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

func (sl *parallelStateLogic) ErrorCatchers() []model.ErrorDefinition {
	return sl.state.ErrorDefinitions()
}

func (sl *parallelStateLogic) ID() string {
	return sl.state.ID
}

func (sl *parallelStateLogic) LivingChildren(savedata []byte) []stateChild {

	var err error
	var children = make([]stateChild, 0)

	var logics []multiactionTuple
	err = json.Unmarshal(savedata, &logics)
	if err != nil {
		log.Error(err)
		return children
	}

	for _, logic := range logics {
		if logic.Complete {
			continue
		}
		children = append(children, stateChild{
			Id:   logic.ID,
			Type: logic.Type,
		})
	}

	return children

}

func (sl *parallelStateLogic) Run(ctx context.Context, instance *workflowLogicInstance, savedata, wakedata []byte) (transition *stateTransition, err error) {

	// TODO: logs

	if len(wakedata) == 0 {

		// first part

		logics := make([]multiactionTuple, 0)

		if len(savedata) != 0 {
			err = NewInternalError(errors.New("got unexpected savedata"))
			return
		}

		if len(sl.state.Actions) > maxParallelActions {
			err = NewUncatchableError("direktiv.limits.parallel", "instance aborted for exceeding the maximum number of parallel actions (%d)", maxParallelActions)
			return
		}

		for _, action := range sl.state.Actions {

			var input interface{}

			input, err = instance.JQ(".")
			if err != nil {
				return
			}

			m, ok := input.(map[string]interface{})
			if !ok {
				err = NewInternalError(errors.New("invalid state data"))
				return
			}

			if len(action.Secrets) > 0 {
				instance.Log("Decrypting secrets.")

				s := make(map[string]string)

				for _, name := range action.Secrets {
					var dd []byte
					dd, err = decryptedDataForNS(ctx, instance, instance.namespace, name)
					if err != nil {
						return
					}
					s[name] = string(dd)
				}

				m["secrets"] = s
			}

			input, err = jq(m, action.Input)
			if err != nil {
				return
			}

			var inputData []byte

			inputData, err = json.Marshal(input)
			if err != nil {
				err = NewInternalError(err)
				return
			}

			if action.Function != "" {

				// container

				uid := ksuid.New()
				logics = append(logics, multiactionTuple{
					ID:   uid.String(),
					Type: "isolate",
				})

				var fn *model.FunctionDefinition

				fn, err = sl.workflow.GetFunction(action.Function)
				if err != nil {
					err = NewInternalError(err)
					return
				}

				ar := new(actionRequest)
				ar.ActionID = uid.String()
				ar.Workflow.InstanceID = instance.id
				ar.Workflow.Namespace = instance.namespace
				ar.Workflow.State = sl.state.GetID()
				ar.Workflow.Step = instance.step
				ar.Container.Image = fn.Image
				ar.Container.Cmd = fn.Cmd
				ar.Container.Size = int32(fn.Size)

				// TODO: timeout
				ar.Container.Data = inputData

				// get registries
				ar.Container.Registries, err = getRegistries(instance.engine.server.config, instance.namespace)
				if err != nil {
					return
				}

				err = instance.engine.doActionRequest(ctx, ar)
				if err != nil {
					return
				}

			} else {

				// subflow

				caller := new(subflowCaller)
				caller.InstanceID = instance.id
				caller.State = sl.state.GetID()
				caller.Step = instance.step

				var subflowID string

				subflowID, err = instance.engine.subflowInvoke(caller, instance.rec.InvokedBy, instance.namespace, action.Workflow, inputData)
				if err != nil {
					return
				}

				logics = append(logics, multiactionTuple{
					ID:   subflowID,
					Type: "subflow",
				})

			}

		}

		var data []byte
		data, err = json.Marshal(logics)
		if err != nil {
			err = NewInternalError(err)
			return
		}

		err = instance.Save(ctx, data)
		if err != nil {
			return
		}

		return

	}

	// second part

	results := new(actionResultPayload)
	err = json.Unmarshal(wakedata, results)
	if err != nil {
		err = NewInternalError(err)
		return
	}

	var logics []multiactionTuple
	err = json.Unmarshal(savedata, &logics)
	if err != nil {
		err = NewInternalError(err)
		return
	}

	var found bool
	var idx int
	var completed int

	for i, lid := range logics {

		if lid.ID == results.ActionID {
			found = true
			if lid.Complete {
				err = NewInternalError(fmt.Errorf("action '%s' already completed", lid.ID))
				return
			}
			logics[i].Complete = true
			lid.Complete = true
			idx = i
		}

		if lid.Complete {
			completed++
		}

	}

	if !found {
		err = NewInternalError(fmt.Errorf("action '%s' wasn't expected", results.ActionID))
		return
	}

	var x interface{}
	err = json.Unmarshal(results.Output, &x)
	if err != nil {
		x = base64.StdEncoding.EncodeToString(results.Output)
	}

	logics[idx].Results = x

	var ready bool
	switch sl.state.Mode {
	case model.BranchModeAnd:

		if completed == len(logics) {
			ready = true
		}

		if results.ErrorCode != "" {
			err = NewCatchableError(results.ErrorCode, results.ErrorMessage)
			return
		}

		if results.ErrorMessage != "" {
			err = NewInternalError(errors.New(results.ErrorMessage))
			return
		}

		/*
			// cancel other unfinished branches
			for _, logic := range logics {
				if logic.Complete {
					continue
				}
				switch logic.Type {
				case "subflow":
					err = syncServer(ctx, instance.engine.db, &instance.engine.server.id, logic.ID, cancelSubflow)
					if err != nil {
						log.Errorf("failed to cancel subflow: %v", err)
					}
				case "isolate":
					err = syncServer(ctx, instance.engine.db, &instance.engine.server.id, logic.ID, cancelIsolate)
					if err != nil {
						log.Errorf("failed to cancel isolate: %v", err)
					}
				default:
					log.Errorf("Unknown logic type: %s", logic.Type)
				}
			}
		*/

	case model.BranchModeOr:

		if results.ErrorCode != "" {
			instance.Log("Branch %d failed with error '%s': %s", idx, results.ErrorCode, results.ErrorMessage)
		} else if results.ErrorMessage != "" {
			instance.Log("Branch %d failed with an internal error: %s", idx, results.ErrorMessage)
		} else {
			ready = true
		}

		if completed == len(logics) {
			err = NewCatchableError(ErrCodeAllBranchesFailed, "all branches failed")
			return
		}

	default:
		err = NewInternalError(errors.New("unrecognized branch mode"))
		return
	}

	if ready {

		var results []interface{}
		for i := range logics {
			results = append(results, logics[i].Results)
		}

		err = instance.StoreData("return", results)
		if err != nil {
			err = NewInternalError(err)
			return
		}

		transition = &stateTransition{
			Transform: sl.state.Transform,
			NextState: sl.state.Transition,
		}

		return

	}

	var data []byte
	data, err = json.Marshal(logics)
	if err != nil {
		err = NewInternalError(err)
		return
	}

	err = instance.Save(ctx, data)
	if err != nil {
		return
	}

	return

}

// -------------- Switch State --------------

type switchStateLogic struct {
	state *model.SwitchState
}

func initSwitchStateLogic(wf *model.Workflow, state model.State) (stateLogic, error) {

	switchState, ok := state.(*model.SwitchState)
	if !ok {
		return nil, NewInternalError(errors.New("bad state object"))
	}

	sl := new(switchStateLogic)
	sl.state = switchState
	return sl, nil

}

func (sl *switchStateLogic) Type() string {
	return model.StateTypeSwitch.String()
}

func (sl *switchStateLogic) Deadline() time.Time {
	return time.Now().Add(time.Second * 5)
}

func (sl *switchStateLogic) ErrorCatchers() []model.ErrorDefinition {
	return sl.state.ErrorDefinitions()
}

func (sl *switchStateLogic) ID() string {
	return sl.state.GetID()
}

func (sl *switchStateLogic) LivingChildren(savedata []byte) []stateChild {
	return nil
}

func (sl *switchStateLogic) Run(ctx context.Context, instance *workflowLogicInstance, savedata, wakedata []byte) (transition *stateTransition, err error) {

	if len(savedata) != 0 {
		err = NewInternalError(errors.New("got unexpected savedata"))
		return
	}

	if len(wakedata) != 0 {
		err = NewInternalError(errors.New("got unexpected wakedata"))
		return
	}

	var success bool

	for i, condition := range sl.state.Conditions {

		var x interface{}
		x, err = instance.JQ(condition.Condition)
		if err != nil {
			err = NewInternalError(fmt.Errorf("switch condition %d condition failed to run: %v", i, err))
			return
		}

		if x != nil {
			switch x.(type) {
			case bool:
				if x.(bool) {
					success = true
				}
			case string:
				if x.(string) != "" {
					success = true
				}
			case int:
				if x.(int) != 0 {
					success = true
				}
			case []interface{}:
				if len(x.([]interface{})) > 0 {
					success = true
				}
			case map[string]interface{}:
				if len(x.(map[string]interface{})) > 0 {
					success = true
				}
			default:
			}
		}

		if success {
			instance.Log("Switch condition %d succeeded", i)
			transition = &stateTransition{
				Transform: condition.Transform,
				NextState: condition.Transition,
			}
			break
		}

	}

	if !success {
		instance.Log("No switch conditions succeeded")
		transition = &stateTransition{
			Transform: sl.state.DefaultTransform,
			NextState: sl.state.DefaultTransition,
		}
	}

	return

}

// -------------- Validate State --------------

type validateStateLogic struct {
	state *model.ValidateState
}

func initValidateStateLogic(wf *model.Workflow, state model.State) (stateLogic, error) {

	validate, ok := state.(*model.ValidateState)
	if !ok {
		return nil, NewInternalError(errors.New("bad state object"))
	}

	sl := new(validateStateLogic)
	sl.state = validate
	return sl, nil

}

func (sl *validateStateLogic) Type() string {
	return model.StateTypeValidate.String()
}

func (sl *validateStateLogic) Deadline() time.Time {
	return time.Now().Add(time.Second * 5)
}

func (sl *validateStateLogic) ErrorCatchers() []model.ErrorDefinition {
	return sl.state.ErrorDefinitions()
}

func (sl *validateStateLogic) ID() string {
	return sl.state.GetID()
}

func (sl *validateStateLogic) LivingChildren(savedata []byte) []stateChild {
	return nil
}

func (sl *validateStateLogic) Run(ctx context.Context, instance *workflowLogicInstance, savedata, wakedata []byte) (transition *stateTransition, err error) {

	if len(savedata) != 0 {
		err = NewInternalError(errors.New("got unexpected savedata"))
		return
	}

	if len(wakedata) != 0 {
		err = NewInternalError(errors.New("got unexpected wakedata"))
		return
	}

	var schemaData []byte
	schemaData, err = json.Marshal(sl.state.Schema)
	if err != nil {
		err = NewInternalError(err)
		return
	}

	subjectQuery := "."
	if sl.state.Subject != "" {
		subjectQuery = sl.state.Subject
	}

	var subject interface{}
	subject, err = instance.JQ(subjectQuery)
	if err != nil {
		return
	}

	documentData, err := json.Marshal(subject)
	if err != nil {
		err = NewInternalError(err)
		return
	}

	schema := gojsonschema.NewStringLoader(string(schemaData))
	document := gojsonschema.NewStringLoader(string(documentData))
	result, err := gojsonschema.Validate(schema, document)
	if err != nil {
		err = NewInternalError(err)
		return
	}

	if !result.Valid() {
		for _, reason := range result.Errors() {
			instance.Log("Schema validation error: %s", reason.String())
		}
		err = NewCatchableError("direktiv.schema.failed", fmt.Sprintf("subject failed its JSONSchema validation: %v", err))
		return
	}

	transition = &stateTransition{
		Transform: sl.state.Transform,
		NextState: sl.state.Transition,
	}

	return

}
