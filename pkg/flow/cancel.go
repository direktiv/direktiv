package flow

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/flow/states"
	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
	"github.com/direktiv/direktiv/pkg/refactor/pubsub"
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

type cancelWorkflowMessage struct {
	ID      string
	Code    string
	Message string
	Soft    bool
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
			x := &cancelWorkflowMessage{
				ID:      child.Id,
				Code:    ErrCodeCancelledByParent,
				Message: "cancelled by parent workflow",
				Soft:    false,
			}
			data := bytedata.Marshal(x)
			engine.pubsub.Publish(pubsub.EngineCancelInstance, data)
		default:
			engine.sugar.Errorf("unrecognized child type: %s", child.Type)
		}
	}
}

func (engine *engine) cancelInstance(id, code, message string, soft bool) {
	engine.cancelRunning(id)

	ctx, im, err := engine.loadInstanceMemory(id, -1)
	if err != nil {
		engine.sugar.Error(err)
		return
	}

	if im.instance.Instance.Status != instancestore.InstanceStatusPending {
		return
	}

	if soft {
		err = derrors.NewCatchableError(code, message)
	} else {
		err = derrors.NewUncatchableError(code, message)
	}

	engine.sugar.Debugf("Handling cancel instance: %s", this())

	go engine.runState(ctx, im, nil, err)
}

func (engine *engine) finishCancelWorkflow(data string) {
	args := new(cancelWorkflowMessage)

	err := json.Unmarshal([]byte(data), &args)
	if err != nil {
		engine.sugar.Error(err)
		return
	}

	engine.cancelInstance(args.ID, args.Code, args.Message, args.Soft)
}

func (engine *engine) cancelRunning(id string) {
	engine.pubsub.Publish(pubsub.TimerDelete, id)

	engine.cancellersLock.Lock()
	cancel, exists := engine.cancellers[id]
	if exists {
		cancel()
	}
	engine.cancellersLock.Unlock()
}

func (engine *engine) finishCancelMirrorProcess(data string) {
	args := make([]interface{}, 0)

	err := json.Unmarshal([]byte(data), &args)
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
