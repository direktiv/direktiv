package flow

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/direktiv/direktiv/pkg/flow/pubsub"
	"github.com/direktiv/direktiv/pkg/flow/states"
	"github.com/google/uuid"
)

func (engine *engine) Children(ctx context.Context, im *instanceMemory) ([]*states.ChildInfo, error) {
	var err error

	var children []*states.ChildInfo
	err = im.UnmarshalMemory(&children)
	if err != nil {
		return nil, err
	}

	return children, nil
}

func (engine *engine) LivingChildren(ctx context.Context, im *instanceMemory) []stateChild {
	living := make([]stateChild, 0)

	children, err := engine.Children(ctx, im)
	if err != nil {
		engine.sugar.Error(err)
		return living
	}

	for _, logic := range children {
		if logic == nil {
			continue
		}
		if logic.Complete {
			continue
		}
		living = append(living, stateChild{
			Id:          logic.ID,
			Type:        logic.Type,
			ServiceName: logic.ServiceName,
		})
	}

	return living
}

func (engine *engine) CancelInstanceChildren(ctx context.Context, im *instanceMemory) {
	children := engine.LivingChildren(ctx, im)

	for _, child := range children {
		switch child.Type {
		case "isolate":
			if child.ServiceName != "" {
				// TODO: yassir handle workflow children services.
			} else {
				engine.sugar.Warn("missing child service name")
			}
		case "subflow":
			engine.pubsub.CancelWorkflow(child.Id, ErrCodeCancelledByParent, "cancelled by parent workflow", false)
		default:
			slog.Error("unrecognized child type", "error", child.Type)
		}
	}
}

func (engine *engine) cancelInstance(id, code, message string, soft bool) {
	ctx := context.Background()

	uid, err := uuid.Parse(id)
	if err != nil {
		slog.Error("failed to parse instance uuid for cancelInstance", "error", err.Error())
		return
	}

	err = engine.enqueueInstanceMessage(ctx, uid, "cancel", cancelMessage{
		Soft:    soft,
		Code:    code,
		Message: message,
	})
	if err != nil {
		slog.Error("failed to enqueue instance message for cancelInstance", "error", err.Error())
		return
	}

	if !soft {
		engine.cancelRunning(id)
	}
}

func (engine *engine) finishCancelWorkflow(req *pubsub.PubsubUpdate) {
	args := make([]interface{}, 0)

	err := json.Unmarshal([]byte(req.Key), &args)
	if err != nil {
		engine.sugar.Error(err)
		return
	}

	var soft, ok bool
	var id, code, msg string

	if len(args) != 4 {
		goto bad
	}

	id, ok = args[0].(string)
	if !ok {
		goto bad
	}

	code, ok = args[1].(string)
	if !ok {
		goto bad
	}

	msg, ok = args[2].(string)
	if !ok {
		goto bad
	}

	soft, ok = args[3].(bool)
	if !ok {
		goto bad
	}

	engine.cancelInstance(id, code, msg, soft)

	return

bad:

	engine.sugar.Error(errors.New("bad input to workflow cancel pubsub"))
}

func (engine *engine) cancelRunning(id string) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return
	}

	engine.sendCancelToScheduled(uid)
}

func (engine *engine) finishCancelMirrorProcess(req *pubsub.PubsubUpdate) {
	args := make([]interface{}, 0)

	err := json.Unmarshal([]byte(req.Key), &args)
	if err != nil {
		engine.sugar.Error(err)
		return
	}

	var ok bool
	var id string
	var uid uuid.UUID

	if len(args) != 1 {
		goto bad
	}

	id, ok = args[0].(string)
	if !ok {
		goto bad
	}

	uid, err = uuid.Parse(id)
	if err != nil {
		goto bad
	}

	_ = engine.mirrorManager.Cancel(context.Background(), uid)

	return

bad:

	engine.sugar.Error(errors.New("bad input to mirror process cancel pubsub"))
}
