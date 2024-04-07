package flow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
	"github.com/google/uuid"
)

func (engine *engine) registerScheduled(ctx context.Context, instance uuid.UUID) (context.Context, error) {
	ctx, cancel := context.WithCancel(ctx)

	_, loaded := engine.scheduled.LoadOrStore(instance, cancel)
	if loaded {
		cancel()
		return nil, fmt.Errorf("cannot register for scheduling because instance is already scheduled")
	}

	return ctx, nil
}

func (engine *engine) deregisterScheduled(instance uuid.UUID) {
	engine.sendCancelToScheduled(instance)
	engine.scheduled.Delete(instance)
}

func (engine *engine) sendCancelToScheduled(instance uuid.UUID) {
	v, ok := engine.scheduled.Load(instance)
	if !ok {
		return
	}

	cancel, ok := v.(func())
	if !ok {
		return
	}

	cancel()
}

func (engine *engine) executor(ctx context.Context, id uuid.UUID) {
	ctx, err := engine.registerScheduled(ctx, id)
	if err != nil {
		slog.Error("Failed to register instance for scheduled execution.", "instance", id, "error", err)
		return
	}
	slog.Debug("Successfully registered instance for scheduled execution.", "instance", id)
	im, err := engine.getInstanceMemory(ctx, id)
	if err != nil {
		slog.Error("Failed to retrieve instance memory in executor.", "instance", id, "error", err)
		engine.deregisterScheduled(id)

		return
	}
	slog.Debug("Beginning instance execution loop.", "instance", id)

	engine.executorLoop(ctx, im)
	slog.Debug("Successfully deregistered instance after execution.", "instance", id)

	engine.deregisterScheduled(id)
}

func (engine *engine) transitionLoop(ctx context.Context, im *instanceMemory, msg *instancestore.InstanceMessageData) {
	for {
		transition := engine.handleInstanceMessage(ctx, im, msg)
		if transition == nil {
			return
		}

		payload, err := json.Marshal(map[string]interface{}{
			"type": "transition",
			"data": transition.NextState,
		})
		if err != nil {
			panic(err)
		}

		msg = &instancestore.InstanceMessageData{
			ID:         uuid.New(),
			InstanceID: im.instance.Instance.ID,
			CreatedAt:  time.Now(),
			Payload:    payload,
		}
	}
}

func (engine *engine) executorLoop(ctx context.Context, im *instanceMemory) {
	for {
		// pop message
		tx, err := engine.flow.beginSqlTx(ctx)
		if err != nil {
			engine.CrashInstance(ctx, im, derrors.NewUncatchableError("", err.Error()))
			return
		}
		defer tx.Rollback()

		msg, err := tx.InstanceStore().ForInstanceID(im.ID()).PopMessage(ctx)
		if err != nil {
			if errors.Is(err, instancestore.ErrNoMessages) {
				// yield
				return
			}

			engine.CrashInstance(ctx, im, derrors.NewUncatchableError("", err.Error()))
			return
		}

		err = tx.Commit(ctx)
		if err != nil {
			engine.CrashInstance(ctx, im, derrors.NewUncatchableError("", err.Error()))
			return
		}

		engine.transitionLoop(ctx, im, msg)
	}
}

func (engine *engine) InstanceYield(ctx context.Context, im *instanceMemory) {
	slog.Debug("Instance preparing to yield and release resources.", "instance", im.ID().String(), "namespace", im.Namespace())

	err := engine.freeMemory(ctx, im)
	if err != nil {
		slog.Error("Failed to free memory for instance. Initiating crash sequence.", "instance", im.ID().String(), "namespace", im.Namespace(), "error", err)
		engine.CrashInstance(ctx, im, err)
		return
	}
}

func (engine *engine) WakeInstanceCaller(ctx context.Context, im *instanceMemory) {
	caller := engine.InstanceCaller(im)

	if caller != nil {
		slog.Debug("Initiating result report to calling workflow.", "namespace", im.Namespace(), "instance", im.ID())

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
			slog.Error("Failed to report action results to caller workflow.", "namespace", im.Namespace(), "instance", im.ID(), "error", err)
			return
		}
	}
}

func (engine *engine) start(im *instanceMemory) {
	namespace := im.instance.TelemetryInfo.NamespaceName
	workflowPath := GetInodePath(im.instance.Instance.WorkflowPath)

	metricsWfInvoked.WithLabelValues(namespace, workflowPath, namespace).Inc()
	metricsWfPending.WithLabelValues(namespace, workflowPath, namespace).Inc()

	ctx := context.Background()

	slog.Debug("Workflow execution initiated.", "namespace", namespace, "workflow", workflowPath, "instance", im.ID())

	workflow, err := im.Model()
	if err != nil {
		engine.CrashInstance(ctx, im, derrors.NewUncatchableError(ErrCodeWorkflowUnparsable, "failed to parse workflow YAML: %v", err))
		slog.Error("Failed to parse workflow YAML. Workflow execution halted.", "namespace", namespace, "workflow", workflowPath, "instance", im.ID(), "error", err)
		return
	}

	//

	id := im.instance.Instance.ID

	ctx, err = engine.registerScheduled(ctx, id)
	if err != nil {
		slog.Debug("Failed to register workflow as scheduled. Workflow execution may be delayed or halted.", "namespace", namespace, "workflow", workflowPath, "instance", id, "error", err)
		return
	}

	payload, err := json.Marshal(map[string]interface{}{
		"type": "transition",
		"data": workflow.GetStartState().GetID(),
	})
	if err != nil {
		slog.Error("Failed to marshal start state payload. Halting workflow execution.", "namespace", namespace, "workflow", workflowPath, "instance", id, "error", err)
		panic(err) // TODO?
	}

	engine.transitionLoop(ctx, im, &instancestore.InstanceMessageData{
		ID:         uuid.New(),
		InstanceID: id,
		CreatedAt:  time.Now(),
		Payload:    payload,
	})

	engine.deregisterScheduled(id)
}
