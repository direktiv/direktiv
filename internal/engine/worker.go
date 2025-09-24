package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	intNats "github.com/direktiv/direktiv/internal/nats"
	"github.com/direktiv/direktiv/pkg/lifecycle"
	"github.com/nats-io/nats.go"
)

const workersCount = 5

func (e *Engine) startQueueWorkers(lc *lifecycle.Manager) error {
	// Bind to the existing durable consumer
	sub, err := e.js.PullSubscribe(
		fmt.Sprintf(intNats.SubjEngineQueue, "*", "*"),
		intNats.ConsumerEngineQueue,
		nats.BindStream(intNats.StreamEngineQueue),
		nats.ManualAck(),
	)
	if err != nil {
		return fmt.Errorf("nats pull subscribe %s: %w", intNats.ConsumerEngineQueue, err)
	}

	for i := range workersCount {
		lc.Go(func() error {
			err := e.runLoop(lc, sub)
			if err != nil {
				return fmt.Errorf("runLoop(%d), err: %w", i, err)
			}

			return nil
		})
	}

	return nil
}

func (e *Engine) runLoop(lc *lifecycle.Manager, sub *nats.Subscription) error {
	for {
		select {
		case <-lc.Done():
			return nil
		default:
		}
		msgList, err := sub.Fetch(1, nats.MaxWait(1*time.Second))
		if err != nil && !errors.Is(err, nats.ErrTimeout) {
			slog.Error("subscriber fetch", "error", err, "subject", sub.Subject)
			continue
		}
		for _, msg := range msgList {
			if err := e.handleQueueMessage(lc.Context(), msg); err != nil {
				slog.Error("handle queue message", "error", err, "msg", string(msg.Data))
				_ = msg.Nak()
			} else {
				_ = msg.Ack()
			}
		}
	}
}

func decodeQueueMessage(msg *nats.Msg) (*InstanceEvent, error) {
	var ev InstanceEvent
	if err := json.Unmarshal(msg.Data, &ev); err != nil {
		return nil, err
	}
	meta, err := msg.Metadata()
	if err != nil {
		return nil, err
	}
	ev.Sequence = meta.Sequence.Stream

	return &ev, nil
}

func (e *Engine) handleQueueMessage(ctx context.Context, msg *nats.Msg) error {
	ev, err := decodeQueueMessage(msg)
	if err != nil {
		return fmt.Errorf("decode queue msg: %w", err)
	}

	err = e.ExecInstance(ctx, ev)
	if err != nil {
		return fmt.Errorf("exec instance: %w", err)
	}

	return nil
}
