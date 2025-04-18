package flow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	enginerefactor "github.com/direktiv/direktiv/pkg/engine"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/flow/states"
	"github.com/direktiv/direktiv/pkg/instancestore"
	"github.com/direktiv/direktiv/pkg/pubsub"
	"github.com/direktiv/direktiv/pkg/telemetry"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

const (
	engineSchedulingTimeout = time.Second * 10
	engineOwnershipTimeout  = time.Minute
)

type instanceMessageChannelData struct {
	InstanceID        uuid.UUID
	LastKnownServer   uuid.UUID
	LastKnownUpdateAt time.Time
}

func (engine *engine) enqueueInstanceMessage(ctx context.Context, id uuid.UUID, kind string, data interface{}) error {
	// TODO: should this add state and step data? At some point these fields died so I have removed them.
	payload, err := json.Marshal(map[string]interface{}{
		"type": kind,
		"data": data,
	})
	if err != nil {
		panic(err)
	}

	// NOTE: we don't do serializable here. We don't need to. This is a best effort logic.
	tx, err := engine.flow.beginSQLTx(ctx) /*&sql.TxOptions{
		Isolation: sql.LevelSerializable,
	}*/if err != nil {
		slog.Error("failed to begin SQL transaction", "error", err)

		return err
	}
	defer tx.Rollback()

	idata, err := tx.InstanceStore().ForInstanceID(id).GetSummary(ctx)
	if err != nil {
		return err
	}

	if idata.EndedAt != nil && !idata.EndedAt.IsZero() {
		return errors.New("message rejected: instance has already ended")
	}

	err = tx.InstanceStore().ForInstanceID(id).EnqueueMessage(ctx, &instancestore.EnqueueInstanceMessageArgs{
		InstanceID: id,
		Payload:    payload,
	})
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	msg, err := json.Marshal(instanceMessageChannelData{
		InstanceID:        id,
		LastKnownServer:   idata.Server,
		LastKnownUpdateAt: idata.UpdatedAt,
	})
	if err != nil {
		panic(err)
	}

	if idata.Server == engine.ID && time.Now().Add(-engineOwnershipTimeout).Before(idata.UpdatedAt) {
		go engine.instanceMessagesChannelHandler(string(msg)) //nolint:contextcheck
	} else {
		err = engine.Bus.Publish(&pubsub.InstanceMessageEvent{
			Message: string(msg),
		})
		if err != nil {
			slog.Error("failed to publish message to bus", "error", err)

			return err
		}
	}

	return nil
}

func (engine *engine) instanceMessagesChannelHandler(data string) {
	var args instanceMessageChannelData

	err := json.Unmarshal([]byte(data), &args)
	if err != nil {
		slog.Error("instanceMessagesChannelHandler failed to unmarshal channel data", "error", err)
		return
	}

	if engine.ID != args.LastKnownServer {
		// if we aren't the last known server, only attempt to take control if it has been a while
		if time.Now().Add(-1 * engineOwnershipTimeout / 2).Before(args.LastKnownUpdateAt) {
			return
		}
	}

	go engine.executor(context.Background(), args.InstanceID)
}

func (engine *engine) handleInstanceMessage(ctx context.Context, im *instanceMemory, msg *instancestore.InstanceMessageData) *states.Transition {
	ctx, span := enginerefactor.TraceGet(ctx, im.instance.TelemetryInfo)

	ctx = im.Context(ctx)

	if im.instance.Instance.EndedAt != nil && !im.instance.Instance.EndedAt.IsZero() {
		telemetry.LogInstance(ctx, telemetry.LogLevelDebug,
			"skipping message because instance has ended")

		return nil
	}

	m := make(map[string]interface{})

	err := json.Unmarshal(msg.Payload, &m)
	if err != nil {
		telemetry.LogInstanceError(ctx, "failed to unmarshal message payload", err)

		return nil
	}

	var ok bool
	var msgType string

	x, ok := m["type"]
	if !ok {
		telemetry.LogInstanceError(ctx, "invalid message payload: missing 'type' field", err)

		return nil
	}

	msgType, ok = x.(string)
	if !ok {
		telemetry.LogInstanceError(ctx, "failed to unmarshal message payload: 'type' field not a string", err)

		return nil
	}

	x, ok = m["data"]
	if !ok {
		telemetry.LogInstanceError(ctx, "invalid message payload: missing 'data' field", err)

		return nil
	}

	span.SetAttributes(attribute.KeyValue{
		Key:   "message",
		Value: attribute.StringValue(msgType),
	})

	span.AddEvent(fmt.Sprintf("received %s", msgType))

	data, _ := json.Marshal(x)
	// TODO trace each of thos branches with spans
	switch msgType {
	case "cancel":
		return engine.handleCancelMessage(ctx, im, data)
	case "wake":
		return engine.handleWakeMessage(ctx, im, data)
	case "event":
		return engine.handleEventMessage(ctx, im, data)
	case "action":
		return engine.handleActionMessage(ctx, im, data)
	case "transition":
		return engine.handleTransitionMessage(ctx, im, data)
	default:
		slog.Error("encountered unrecognized instance message type", "msgType", msgType, "instance", im.ID(), "namespace", im.Namespace().Name)

		panic(fmt.Sprintf("unrecognized instance message type: %s", msgType))
	}
}

type cancelMessage struct {
	Code    string
	Message string
	Soft    bool
}

func (engine *engine) handleCancelMessage(ctx context.Context, im *instanceMemory, data []byte) *states.Transition {
	var args cancelMessage

	err := json.Unmarshal(data, &args)
	if err != nil {
		slog.Error("handleCancelMessage failed to unmarshal cancel message args", "error", err)
		return nil
	}

	if args.Soft {
		err = derrors.NewCatchableError(args.Code, "%s", args.Message)
	} else {
		err = derrors.NewUncatchableError(args.Code, "%s", args.Message)
	}

	return engine.runState(ctx, im, nil, err)
}

func (engine *engine) handleWakeMessage(ctx context.Context, im *instanceMemory, data []byte) *states.Transition {
	var pl retryMessage

	ctx = im.Context(ctx)

	err := json.Unmarshal(data, &pl)
	if err != nil {
		telemetry.LogInstanceError(ctx, "handleWakeMessage failed to unmarshal wakeup message args", err)

		return nil
	}

	return engine.runState(ctx, im, pl.Data, nil)
}

func (engine *engine) handleActionMessage(ctx context.Context, im *instanceMemory, data []byte) *states.Transition {
	var pl actionResultPayload

	err := json.Unmarshal(data, &pl)
	if err != nil {
		telemetry.LogInstanceError(ctx, "handleActionMessage failed to unmarshal action results message", err)

		return nil
	}

	return engine.runState(ctx, im, data, nil)
}

func (engine *engine) handleEventMessage(ctx context.Context, im *instanceMemory, data []byte) *states.Transition {
	ctx, span := telemetry.Tracer.Start(ctx, "event-received")
	defer span.End()
	span.AddEvent("cloud-event received")
	span.SetAttributes(attribute.KeyValue{
		Key:   "event",
		Value: attribute.StringValue(string(data)),
	})

	return engine.runState(ctx, im, data, nil)
}

func (engine *engine) handleTransitionMessage(ctx context.Context, im *instanceMemory, data []byte) *states.Transition {
	var state string

	err := json.Unmarshal(data, &state)
	if err != nil {
		telemetry.LogInstanceError(ctx, "handleTransitionMessage failed to unmarshal transition message args", err)

		return nil
	}

	// this is the main loop and we have to update the context for tracing
	// the loop runs in the root context but for the transition we
	// have to set it from the parent state
	ctx = telemetry.FromTraceParent(ctx, im.instance.TelemetryInfo.TraceParent)

	return engine.Transition(ctx, im, state, 0)
}
