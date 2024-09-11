package flow

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

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
			ID:          logic.ID,
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
				slog.Warn("Isolate child missing service name.", "child_id", child.ID)
			}
		case "subflow":
			err := engine.pBus.Publish(cancelInstanceMessage{
				ID:      child.ID,
				Code:    ErrCodeCancelledByParent,
				Message: "cancelled by parent workflow",
				Soft:    false,
			})
			if err != nil {
				slog.Error("Publish error", "error", err)
			}
		default:
			slog.Error("Encountered unrecognized child type.", "error", child.Type)
		}
	}

	return nil
}

type cancelInstanceMessage struct {
	ID      string
	Code    string
	Message string
	Soft    bool
}

func (engine *engine) cancelInstanceHandler(data string) {
	var msg cancelInstanceMessage
	err := json.Unmarshal([]byte(data), &msg)
	if err != nil {
		slog.Error(err.Error())

		return
	}

	engine.cancelInstance(msg.ID, msg.Code, msg.Message, msg.Soft)
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

func (engine *engine) cancelRunning(id string) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return
	}

	engine.sendCancelToScheduled(uid)
}
