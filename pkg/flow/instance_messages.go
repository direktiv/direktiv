package flow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/flow/states"
	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
	"github.com/google/uuid"
)

const (
	engineInstanceMessagesChannel = "instance_messages"

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
	tx, err := engine.flow.beginSqlTx(ctx) /*&sql.TxOptions{
		Isolation: sql.LevelSerializable,
	}*/if err != nil {
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
		go engine.instanceMessagesChannelHandler(string(msg))
	} else {
		err = engine.pBus.Publish(engineInstanceMessagesChannel, string(msg))
		if err != nil {
			return err
		}
	}

	return nil
}

func (engine *engine) instanceMessagesChannelHandler(data string) {
	var args instanceMessageChannelData

	err := json.Unmarshal([]byte(data), &args)
	if err != nil {
		engine.sugar.Errorf("instanceMessagesChannelHandler failed to unmarshal channel data %v", err)
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
	if im.instance.Instance.EndedAt != nil && !im.instance.Instance.EndedAt.IsZero() {
		engine.sugar.Warn("handleInstanceMessage skipping message because instance has ended")
		return nil
	}

	m := make(map[string]interface{})

	err := json.Unmarshal(msg.Payload, &m)
	if err != nil {
		engine.sugar.Errorf("handleInstanceMessage failed to unmarshal message payload: %v", err)
		return nil
	}

	var ok bool
	var msgType string

	x, ok := m["type"]
	if !ok {
		engine.sugar.Errorf("handleInstanceMessage failed to unmarshal message payload: missing 'type' field")
		return nil
	}

	msgType, ok = x.(string)
	if !ok {
		engine.sugar.Errorf("handleInstanceMessage failed to unmarshal message payload: 'type' field not a string")
		return nil
	}

	x, ok = m["data"]
	if !ok {
		engine.sugar.Errorf("handleInstanceMessage got invalid message payload: missing 'data' field")
		return nil
	}

	//nolint:errchkjson
	data, _ := json.Marshal(x)

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
		engine.sugar.Errorf("handleCancelMessage failed to unmarshal cancel message args: %v", err)
		return nil
	}

	if args.Soft {
		err = derrors.NewCatchableError(args.Code, args.Message)
	} else {
		err = derrors.NewUncatchableError(args.Code, args.Message)
	}

	return engine.runState(ctx, im, nil, err)
}

func (engine *engine) handleWakeMessage(ctx context.Context, im *instanceMemory, data []byte) *states.Transition {
	var pl retryMessage

	err := json.Unmarshal(data, &pl)
	if err != nil {
		engine.sugar.Errorf("handleWakeMessage failed to unmarshal wakeup message args: %v", err)
		return nil
	}

	return engine.runState(ctx, im, pl.Data, nil)
}

func (engine *engine) handleActionMessage(ctx context.Context, im *instanceMemory, data []byte) *states.Transition {
	var pl actionResultPayload

	err := json.Unmarshal(data, &pl)
	if err != nil {
		engine.sugar.Errorf("handleActionMessage failed to unmarshal action results message args: %v", err)
		return nil
	}

	traceActionResult(ctx, &pl)

	return engine.runState(ctx, im, data, nil)
}

func (engine *engine) handleEventMessage(ctx context.Context, im *instanceMemory, data []byte) *states.Transition {
	ctx, cleanup, err := traceStateGenericBegin(ctx, im)
	if err != nil {
		engine.sugar.Errorf("handleEventMessage failed to begin trace: %v", err)
		return nil
	}
	defer cleanup()

	return engine.runState(ctx, im, data, nil)
}

func (engine *engine) handleTransitionMessage(ctx context.Context, im *instanceMemory, data []byte) *states.Transition {
	var state string

	err := json.Unmarshal(data, &state)
	if err != nil {
		engine.sugar.Errorf("handleTransitionMessage failed to unmarshal transition message args: %v", err)
		return nil
	}

	return engine.Transition(ctx, im, state, 0)
}
