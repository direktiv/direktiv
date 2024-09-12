package flow

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"log/slog"
	"math/big"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/core"
	enginerefactor "github.com/direktiv/direktiv/pkg/engine"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/instancestore"
	"github.com/direktiv/direktiv/pkg/tracing"
	"github.com/google/uuid"
)

func (engine *engine) registerScheduled(ctx context.Context, instance uuid.UUID) (context.Context, error) {
	ctx, cancel := context.WithCancel(ctx)

	_, loaded := engine.scheduled.LoadOrStore(instance, cancel)
	if loaded {
		cancel()
		return nil, errEngineSync
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
	var im *instanceMemory

	for attempts := range 3 {
		if attempts > 0 {
			a, _ := rand.Int(rand.Reader, big.NewInt(25)) //nolint
			jitter := time.Millisecond * time.Duration(a.Int64())
			time.Sleep(time.Millisecond*50 + jitter)
		}

		im = nil

		ctx2, err := engine.registerScheduled(ctx, id)
		if err != nil {
			if errors.Is(err, errEngineSync) {
				slog.DebugContext(ctx, "Failed to register instance for scheduled execution.", "instance", id, "error", err)
				continue
			}

			slog.ErrorContext(ctx, "Failed to register instance for scheduled execution.", "instance", id, "error", err)

			return
		}
		slog.DebugContext(ctx, "Successfully registered instance for scheduled execution.", "instance", id)

		im, err = engine.getInstanceMemory(ctx2, id)
		if err != nil {
			if strings.Contains(err.Error(), "could not serialize") ||
				strings.Contains(err.Error(), "database records instance terminated") ||
				errors.Is(err, errEngineSync) {
				slog.Debug("Failed to retrieve instance memory in executor.", "instance", id, "error", err)
				engine.deregisterScheduled(id)

				continue
			}

			slog.ErrorContext(ctx, "Failed to retrieve instance memory in executor.", "instance", id, "error", err)

			engine.deregisterScheduled(id)

			return
		}

		ctx = ctx2

		break
	}

	if im == nil {
		return
	}

	slog.Debug("Beginning instance execution loop.", "instance", id)
	ctx = tracing.AddInstanceMemoryAttr(ctx, tracing.InstanceAttributes{
		Namespace:    im.Namespace().Name,
		InstanceID:   im.GetInstanceID().String(),
		Invoker:      im.instance.Instance.Invoker,
		Callpath:     tracing.CreateCallpath(im.instance),
		WorkflowPath: im.instance.Instance.WorkflowPath,
		Status:       core.LogUnknownStatus,
	}, im.GetState())
	ctx = tracing.WithTrack(ctx, tracing.BuildInstanceTrack(im.instance))
	ctx, span, err2 := tracing.InjectTraceParent(ctx, im.instance.TelemetryInfo.TraceParent, "scheduler continues instance: "+im.instance.Instance.WorkflowPath)
	if err2 != nil {
		slog.Warn("engine executor failed to inject trace parent", "error", err2)
	}
	defer span.End()
	engine.executorLoop(ctx, im)
	slog.DebugContext(ctx, "Successfully deregistered instance after execution.", "instance", id)

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
	ctx, cleanup, err := tracing.NewSpan(ctx, "instance scheduling")
	if err != nil {
		slog.Debug("telemetry failed in scheduler", "error", err)
	}
	defer cleanup()
	for {
		// pop message
		tx, err := engine.flow.beginSQLTx(ctx)
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
	slog.DebugContext(ctx, "Instance preparing to yield and release resources.", "instance", im.ID().String(), "namespace", im.Namespace().Name)

	err := engine.freeMemory(ctx, im)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to free memory for instance. Initiating crash sequence.", "instance", im.ID().String(), "namespace", im.Namespace(), "error", err)
		engine.CrashInstance(ctx, im, err)

		return
	}
}

func (engine *engine) WakeInstanceCaller(ctx context.Context, im *instanceMemory) {
	caller := engine.InstanceCaller(im)

	if caller != nil {
		slog.DebugContext(ctx, "Initiating result report to calling workflow.", "namespace", im.Namespace().Name, "instance", im.ID())
		callpath := im.instance.Instance.ID.String()
		for _, v := range im.instance.DescentInfo.Descent {
			callpath += "/" + v.ID.String()
		}
		msg := &actionResultMessage{
			InstanceID: caller.ID.String(),
			ActionContext: enginerefactor.ActionContext{
				TraceParent: im.instance.TelemetryInfo.TraceParent,
				State:       caller.State,
				Branch:      caller.Branch,
				Callpath:    callpath,
				Instance:    im.GetInstanceID().String(),
				Workflow:    im.instance.Instance.WorkflowPath,
				Namespace:   im.instance.Instance.Namespace,
				Step:        caller.Step,
				Action:      im.ID().String(),
			},
			Payload: actionResultPayload{
				ActionID:     im.ID().String(),
				ErrorCode:    im.ErrorCode(),
				ErrorMessage: im.ErrorMessage(),
				Output:       []byte(im.MarshalOutput()),
			},
		}
		err := engine.ReportActionResults(ctx, msg)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to report action results to caller workflow.", "namespace", im.Namespace(), "instance", im.ID(), "error", err)

			return
		}
	}
}

func (engine *engine) start(im *instanceMemory) {
	namespace := im.instance.TelemetryInfo.NamespaceName
	workflowPath := GetInodePath(im.instance.Instance.WorkflowPath)

	ctx := context.Background()
	ctx, span, err := tracing.InjectTraceParent(ctx, im.instance.TelemetryInfo.TraceParent, "scheduler starts instance: "+im.GetInstanceID().String()+", workflow: "+im.instance.Instance.WorkflowPath)
	if err != nil {
		slog.Debug("Failed to populate tracing information. Workflow execution halted.", "namespace", namespace, "workflow", workflowPath, "instance", im.ID(), "error", err)
	}
	defer span.End()
	slog.Debug("Workflow execution initiated.", "namespace", namespace, "workflow", workflowPath, "instance", im.ID())

	workflow, err := im.Model()
	if err != nil {
		engine.CrashInstance(ctx, im, derrors.NewUncatchableError(ErrCodeWorkflowUnparsable, "failed to parse workflow YAML: %v", err))
		slog.ErrorContext(ctx, "Failed to parse workflow YAML. Workflow execution halted.", "namespace", namespace, "workflow", workflowPath, "instance", im.ID(), "error", err)

		return
	}

	id := im.instance.Instance.ID

	ctx, err = engine.registerScheduled(ctx, id)
	if err != nil {
		slog.DebugContext(ctx, "Failed to register workflow as scheduled. Workflow execution may be delayed or halted.", "namespace", namespace, "workflow", workflowPath, "instance", id, "error", err)

		return
	}

	payload, err := json.Marshal(map[string]interface{}{
		"type": "transition",
		"data": workflow.GetStartState().GetID(),
	})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to marshal start state payload. Halting workflow execution.", "namespace", namespace, "workflow", workflowPath, "instance", id, "error", err)
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

func (engine *engine) ReportActionResults(ctx context.Context, req *actionResultMessage) error {
	payload := &actionResultPayload{
		ActionID:     req.Payload.ActionID,
		ErrorCode:    req.Payload.ErrorCode,
		ErrorMessage: req.Payload.ErrorMessage,
		Output:       req.Payload.Output,
	}

	uid, err := uuid.Parse(req.InstanceID)
	if err != nil {
		slog.DebugContext(ctx, "failed in ReportActionResults", "this", this(), "error", err)
		return err
	}

	err = engine.enqueueInstanceMessage(ctx, uid, "action", payload)
	if err != nil {
		slog.DebugContext(ctx, "failed to enqueque ReportActionResults", "this", this(), "error", err)
		return err
	}

	return nil
}
