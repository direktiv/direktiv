package flow

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
)

func (engine *engine) scheduleTimeout(_ context.Context, im *instanceMemory, oldController string, t time.Time, soft bool) {
	var err error
	deadline := t

	prefixes := []string{"soft", "hard"}
	prefix := prefixes[1]
	if soft {
		prefix = prefixes[0]
	}

	oldID := fmt.Sprintf("timeout:%s:%s:%d", im.ID().String(), prefix, im.Step()-1)
	id := fmt.Sprintf("timeout:%s:%s:%d", im.ID().String(), prefix, im.Step())
	if im.Step() == 0 {
		id = fmt.Sprintf("timeout:%s:%s", im.ID().String(), prefix)
	}

	// cancel existing timeouts
	slog.Debug("Cancelling existing timeouts.", "namespace", im.Namespace(), "instance", im.ID(), "timeout_type", prefix, "step", im.Step(), "error", err)

	engine.timers.deleteTimerByName(oldID)
	engine.timers.deleteTimerByName(id)

	// schedule timeout

	args := &timeoutArgs{
		InstanceID: im.ID().String(),
		Step:       im.Step(),
		Soft:       soft,
	}

	data, err := json.Marshal(args)
	if err != nil {
		panic(err)
	}

	err = engine.timers.addOneShot(id, timeoutFunction, deadline, data)
	if err != nil {
		slog.Error("Failed to schedule a timeout.", "namespace", im.Namespace(), "instance", im.ID(), "timeout_type", prefix, "step", im.Step(), "error", err)
	} else {
		slog.Debug("Successfully scheduled a new timeout.", "namespace", im.Namespace(), "instance", im.ID(), "timeout_type", prefix, "step", im.Step(), "error", err)
	}
}

func (engine *engine) ScheduleHardTimeout(ctx context.Context, im *instanceMemory, oldController string, t time.Time) {
	engine.scheduleTimeout(ctx, im, oldController, t, false)
}

func (engine *engine) ScheduleSoftTimeout(ctx context.Context, im *instanceMemory, oldController string, t time.Time) {
	engine.scheduleTimeout(ctx, im, oldController, t, true)
}

type timeoutArgs struct {
	InstanceID string
	Step       int
	Soft       bool
}

const timeoutFunction = "timeoutFunction"

func (engine *engine) hardCancelInstance(instanceID, code, message string) {
	engine.cancelInstance(instanceID, code, message, false)
}

func (engine *engine) softCancelInstance(instanceID string, code, message string) {
	engine.cancelInstance(instanceID, code, message, true)
}

func (engine *engine) timeoutHandler(input []byte) {
	args := new(timeoutArgs)
	err := json.Unmarshal(input, args)
	if err != nil {
		slog.Error("Failed to unmarshal timeout handler arguments.", "error", err)
		return
	}

	if args.Soft {
		slog.Error("Initiating soft cancellation due to timeout.", "instance", args.InstanceID)
		engine.softCancelInstance(args.InstanceID, ErrCodeSoftTimeout, "operation timed out")
		slog.Error("Soft cancellation complete.", "instance", args.InstanceID)
	} else {
		slog.Error("Initiating hard cancellation due to timeout.", "instance", args.InstanceID)
		engine.hardCancelInstance(args.InstanceID, ErrCodeHardTimeout, "workflow timed out")
		slog.Error("Hard cancellation complete.", "instance", args.InstanceID)
	}
}
