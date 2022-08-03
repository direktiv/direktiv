package flow

import (
	"context"
	"encoding/json"
	"time"

	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
)

func (engine *engine) InstanceYield(im *instanceMemory) {

	engine.sugar.Debugf("Instance going to sleep: %s", im.ID().String())

	engine.freeResources(im)

	if im.lock != nil {
		engine.InstanceUnlock(im)
	}

}

func (engine *engine) WakeInstanceCaller(ctx context.Context, im *instanceMemory) {

	caller := engine.InstanceCaller(ctx, im)

	if caller != nil {

		engine.logToInstance(ctx, time.Now(), im.in, "Reporting results to calling workflow.")

		msg := &actionResultMessage{
			InstanceID: caller.InstanceID,
			State:      caller.State,
			Step:       caller.Step,
			Payload: actionResultPayload{
				ActionID:     im.ID().String(),
				ErrorCode:    im.ErrorCode(),
				ErrorMessage: im.ErrorMessage(),
				Output:       []byte(im.MarshalOutput()),
			},
		}

		step := int32(msg.Step)

		// TODO: TRACE

		_, err := engine.server.internal.ReportActionResults(ctx, &grpc.ReportActionResultsRequest{
			InstanceId:   msg.InstanceID,
			Step:         step,
			ActionId:     msg.Payload.ActionID,
			ErrorCode:    msg.Payload.ErrorCode,
			ErrorMessage: msg.Payload.ErrorMessage,
			Output:       msg.Payload.Output,
		})
		if err != nil {
			engine.sugar.Error(err)
			return
		}

	}

}

const sleepWakeupFunction = "sleepWakeup"
const sleepWakedata = "sleep"

type sleepMessage struct {
	InstanceID string
	State      string
	Step       int
}

func (engine *engine) InstanceSleep(ctx context.Context, im *instanceMemory, state string, t time.Time) error {

	data, err := json.Marshal(&sleepMessage{
		InstanceID: im.ID().String(),
		State:      state,
		Step:       im.Step(),
	})
	if err != nil {
		return derrors.NewInternalError(err)
	}

	err = engine.timers.addOneShot(im.ID().String(), sleepWakeupFunction, t, data)
	if err != nil {
		return derrors.NewInternalError(err)
	}

	return nil

}

func (engine *engine) sleepWakeup(data []byte) {

	msg := new(sleepMessage)

	err := json.Unmarshal(data, msg)
	if err != nil {
		engine.sugar.Errorf("cannot handle sleep wakeup: %v", err)
		return
	}

	ctx, im, err := engine.loadInstanceMemory(msg.InstanceID, msg.Step)
	if err != nil {
		engine.sugar.Errorf("cannot load workflow logic instance: %v", err)
		return
	}

	go engine.runState(ctx, im, []byte(sleepWakedata), nil)

}

func (engine *engine) queue(im *instanceMemory) {

	namespace := im.in.Edges.Namespace.Name
	workflow := GetInodePath(im.in.As)

	metricsWfInvoked.WithLabelValues(namespace, workflow, namespace).Inc()
	metricsWfPending.WithLabelValues(namespace, workflow, namespace).Inc()

	// TODO: expand on this
	go engine.start(im)

}

func (engine *engine) hold(im *instanceMemory) {

}
