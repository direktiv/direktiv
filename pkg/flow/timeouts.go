package flow

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
)

func (engine *engine) scheduleTimeout(im *instanceMemory, oldController string, t time.Time, soft bool) {
	var err error
	deadline := t

	prefixes := []string{"soft", "hard"}
	prefix := prefixes[1]
	if soft {
		prefix = prefixes[0]
	}

	oldId := fmt.Sprintf("timeout:%s:%s:%d", im.ID().String(), prefix, im.Step()-1)
	id := fmt.Sprintf("timeout:%s:%s:%d", im.ID().String(), prefix, im.Step())
	if im.Step() == 0 {
		id = fmt.Sprintf("timeout:%s:%s", im.ID().String(), prefix)
	}

	// cancel existing timeouts

	engine.timers.deleteTimerByName(oldController, engine.pubsub.Hostname, oldId)
	engine.timers.deleteTimerByName(oldController, engine.pubsub.Hostname, id)

	// schedule timeout

	args := &timeoutArgs{
		InstanceId: im.ID().String(),
		Step:       im.Step(),
		Soft:       soft,
	}

	data, err := json.Marshal(args)
	if err != nil {
		panic(err)
	}

	err = engine.timers.addOneShot(id, timeoutFunction, deadline, data)
	if err != nil {
		slog.Error("scheduleTimeout", "error", err)
	}
}

func (engine *engine) ScheduleHardTimeout(im *instanceMemory, oldController string, t time.Time) {
	engine.scheduleTimeout(im, oldController, t, false)
}

func (engine *engine) ScheduleSoftTimeout(im *instanceMemory, oldController string, t time.Time) {
	engine.scheduleTimeout(im, oldController, t, true)
}

type timeoutArgs struct {
	InstanceId string
	Step       int
	Soft       bool
}

const timeoutFunction = "timeoutFunction"

func (engine *engine) hardCancelInstance(instanceId, code, message string) {
	engine.cancelInstance(instanceId, code, message, false)
}

func (engine *engine) softCancelInstance(instanceId string, code, message string) {
	engine.cancelInstance(instanceId, code, message, true)
}

func (engine *engine) timeoutHandler(input []byte) {
	args := new(timeoutArgs)
	err := json.Unmarshal(input, args)
	if err != nil {
		slog.Error("timeoutHandler", "error", err)
		return
	}

	if args.Soft {
		engine.softCancelInstance(args.InstanceId, ErrCodeSoftTimeout, "operation timed out")
	} else {
		engine.hardCancelInstance(args.InstanceId, ErrCodeHardTimeout, "workflow timed out")
	}
}
