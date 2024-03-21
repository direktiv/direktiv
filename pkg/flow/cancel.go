package flow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

func (engine *engine) LivingChildren(ctx context.Context, im *instanceMemory) ([]stateChild, error) {
	living := make([]stateChild, 0)

	children, err := engine.Children(ctx, im)
	if err != nil {
		return living, err
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

	return living, nil
}

func (engine *engine) CancelInstanceChildren(ctx context.Context, im *instanceMemory) error {
	children, err := engine.LivingChildren(ctx, im)
	if err != nil {
		return fmt.Errorf("canceling a child failed %w", err)
	}
	for _, child := range children {
		switch child.Type {
		case "isolate":
			if child.ServiceName != "" {
				// TODO: yassir handle workflow children services.
			} else {
				slog.Warn("Isolate child missing service name.", "child_id", child.Id)
			}
		case "subflow":
			engine.pubsub.CancelWorkflow(child.Id, ErrCodeCancelledByParent, "cancelled by parent workflow", false)
		default:
			slog.Error("Encountered unrecognized child type.", "error", child.Type)
		}
	}
	return nil
}

func (engine *engine) cancelInstance(id, code, message string, soft bool) {
	ctx := context.Background()

	uid, err := uuid.Parse(id)
	if err != nil {
		slog.Error("Failed to parse instance UUID.", "error", err)
		return
	}

	err = engine.enqueueInstanceMessage(ctx, uid, "cancel", cancelMessage{
		Soft:    soft,
		Code:    code,
		Message: message,
	})
	if err != nil {
		slog.Error("Failed to enqueue cancel instance message", "error", err, "instance", uid)
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
		slog.Error("finishCancelWorkflow", "error", err)
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

	slog.Error("cancel a workflow", "error", errors.New("bad input to workflow cancel pubsub"))
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
		slog.Error("cancel mirror", "error", err)
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

	slog.Error("mirror process", "error", errors.New("bad input to mirror process cancel pubsub"))
}
