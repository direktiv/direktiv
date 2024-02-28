package flow

import (
	"context"
	"log/slog"

	"github.com/direktiv/direktiv/pkg/flow/grpc"
)

func (engine *engine) InstanceYield(ctx context.Context, im *instanceMemory) {
	engine.sugar.Debugf("Instance going to sleep: %s", im.ID().String())

	e := im.flushUpdates(ctx)
	if e != nil {
		engine.CrashInstance(ctx, im, e)
		return
	}

	engine.freeResources(im)

	if im.lock != nil {
		engine.InstanceUnlock(im)
	}
}

func (engine *engine) WakeInstanceCaller(ctx context.Context, im *instanceMemory) {
	caller := engine.InstanceCaller(im)

	if caller != nil {
		engine.logger.Infof(ctx, im.GetInstanceID(), im.GetAttributes(), "Reporting results to calling workflow.")
		slog.Info("Reporting results to calling workflow.", im.GetSlogAttributes(ctx)...)

		msg := &actionResultMessage{
			InstanceID: caller.ID.String(),
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

func (engine *engine) queue(im *instanceMemory) {
	namespace := im.instance.TelemetryInfo.NamespaceName
	workflow := GetInodePath(im.instance.Instance.WorkflowPath)

	metricsWfInvoked.WithLabelValues(namespace, workflow, namespace).Inc()
	metricsWfPending.WithLabelValues(namespace, workflow, namespace).Inc()

	go engine.start(im)
}
