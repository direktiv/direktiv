package flow

import (
	"context"
	"encoding/json"
	"errors"
)

func (engine *engine) CancelInstanceChildren(ctx context.Context, im *instanceMemory) {

	logic := im.logic

	children := logic.LivingChildren(ctx, engine, im)
	for _, child := range children {
		switch child.Type {
		case "isolate":
			// TODO
			// engine.pubsub.CancelFunction(child.Id)
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

	if im.in.Status != StatusPending {
		return
	}

	if soft {
		err = NewCatchableError(code, message)
	} else {
		err = NewUncatchableError(code, message)
	}

	engine.sugar.Debugf("Handling cancel instance: %s", this())

	go engine.runState(ctx, im, nil, err)

}

func (engine *engine) finishCancelWorkflow(req *PubsubUpdate) {

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

	im, err := engine.getInstanceMemory(context.Background(), engine.db.Instance, id)
	if err == nil {
		engine.timers.deleteTimerByName(im.Controller(), engine.pubsub.hostname, id)
	}

	engine.cancellersLock.Lock()
	cancel, exists := engine.cancellers[id]
	if exists {
		delete(engine.cancellers, id)
		defer cancel()
	}
	engine.cancellersLock.Unlock()

}
