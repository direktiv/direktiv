package flow

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/flow/pubsub"
	"github.com/direktiv/direktiv/pkg/flow/states"
	igrpc "github.com/direktiv/direktiv/pkg/functions/grpc"
	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
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
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				_, err := engine.functionsClient.CancelWorfklow(ctx, &igrpc.FunctionsCancelWorkflowRequest{
					ServiceName: &child.ServiceName,
					ActionID:    &child.Id,
				})
				if err != nil {
					engine.sugar.Errorf("error cancelling action: %v", err)
				}
			} else {
				engine.sugar.Warn("missing child service name")
			}
		case "subflow":
			engine.pubsub.CancelWorkflow(child.Id, ErrCodeCancelledByParent, "cancelled by parent workflow", false)
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
	im, err := engine.getInstanceMemory(context.Background(), id)
	if err == nil {
		engine.timers.deleteTimerByName(im.Controller(), engine.pubsub.Hostname, id)
	}

	engine.cancellersLock.Lock()
	cancel, exists := engine.cancellers[id]
	if exists {
		cancel()
	}
	engine.cancellersLock.Unlock()
}
