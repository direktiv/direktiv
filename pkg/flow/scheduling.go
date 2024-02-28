package flow

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
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
		slog.Error("failed to registerScheduled in executor", "error", err.Error())
		return
	}

	im, err := engine.getInstanceMemory(ctx, id)
	if err != nil {
		slog.Error("failed to getInstanceMemory in executor", "error", err.Error())
		engine.deregisterScheduled(id)

		return
	}

	engine.executorLoop(ctx, im)

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
			engine.CrashInstance(ctx, im, derrors.NewUncatchableError("", err.Error()))
			return
		}

		if msg == nil {
			// yield
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
	engine.sugar.Debugf("Instance going to sleep: %s", im.ID().String())

	err := engine.freeMemory(ctx, im)
	if err != nil {
		engine.CrashInstance(ctx, im, err)
		return
	}
}

func (engine *engine) WakeInstanceCaller(ctx context.Context, im *instanceMemory) {
	caller := engine.InstanceCaller(im)

	if caller != nil {
		engine.logger.Debugf(ctx, im.GetInstanceID(), im.GetAttributes(), "Reporting results to calling workflow.")
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

func (engine *engine) start(im *instanceMemory) {
	namespace := im.instance.TelemetryInfo.NamespaceName
	workflowPath := GetInodePath(im.instance.Instance.WorkflowPath)

	metricsWfInvoked.WithLabelValues(namespace, workflowPath, namespace).Inc()
	metricsWfPending.WithLabelValues(namespace, workflowPath, namespace).Inc()

	ctx := context.Background()

	engine.sugar.Debugf("Starting workflow %v", im.ID().String())
	engine.logger.Debugf(ctx, im.instance.Instance.NamespaceID, im.instance.GetAttributes(recipient.Namespace), "Starting workflow %v", database.GetWorkflow(im.instance.Instance.WorkflowPath))
	slog.Info(fmt.Sprintf("Starting workflow %v", im.instance.Instance.WorkflowPath), "stream", string(recipient.Namespace)+"."+im.Namespace().Name)
	engine.logger.Debugf(ctx, im.instance.Instance.ID, im.GetAttributes(), "Starting workflow %v.", database.GetWorkflow(im.instance.Instance.WorkflowPath))
	slog.Info(fmt.Sprintf("Starting workflow %v.", im.instance.Instance.WorkflowPath), im.instance.GetSlogAttributes(ctx)...)

	workflow, err := im.Model()
	if err != nil {
		engine.CrashInstance(ctx, im, derrors.NewUncatchableError(ErrCodeWorkflowUnparsable, "failed to parse workflow YAML: %v", err))
		engine.logger.Errorf(ctx, im.instance.Instance.NamespaceID, im.instance.GetAttributes(recipient.Namespace), "failed to parse workflow YAML")
		return
	}

	//

	id := im.instance.Instance.ID

	ctx, err = engine.registerScheduled(ctx, id)
	if err != nil {
		slog.Error("failed to registerScheduled in start", "error", err.Error())
		return
	}

	payload, err := json.Marshal(map[string]interface{}{
		"type": "transition",
		"data": workflow.GetStartState().GetID(),
	})
	if err != nil {
		panic(err)
	}

	engine.transitionLoop(ctx, im, &instancestore.InstanceMessageData{
		ID:         uuid.New(),
		InstanceID: id,
		CreatedAt:  time.Now(),
		Payload:    payload,
	})

	engine.deregisterScheduled(id)
}
