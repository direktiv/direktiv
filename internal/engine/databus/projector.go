package databus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/direktiv/direktiv/internal/engine"
	intNats "github.com/direktiv/direktiv/internal/nats"
	"github.com/direktiv/direktiv/pkg/lifecycle"
	"github.com/nats-io/nats.go"
)

const (
	fetchBatch = 200
)

type projector struct {
	js nats.JetStreamContext
}

func (p *projector) start(lc *lifecycle.Manager) error {
	// Bind to the existing durable consumer
	sub, err := intNats.StreamInstanceHistory.PullSubscribe(p.js, nats.ManualAck())
	if err != nil {
		return fmt.Errorf("nats pull subscript to instances.history stream: %w", err)
	}

	lc.Go(func() error {
		err := p.runLoop(lc, sub)
		if err != nil {
			return fmt.Errorf("runLoop, err: %w", err)
		}

		return nil
	})

	return nil
}

func (p *projector) runLoop(lc *lifecycle.Manager, sub *nats.Subscription) error {
	for {
		select {
		case <-lc.Done():
			return nil
		default:
		}
		msgList, err := sub.Fetch(fetchBatch, nats.MaxWait(2*time.Second))
		if err != nil && !errors.Is(err, nats.ErrTimeout) {
			slog.Error("fetch instances.history stream messages", "error", err)
			continue
		}
		for _, msg := range msgList {
			if err := p.handleHistoryMessage(lc.Context(), msg); err != nil {
				slog.Error("handle instances.history stream messages", "error", err)
				_ = msg.Nak()
			} else {
				_ = msg.Ack()
			}
		}
	}
}

func (p *projector) handleHistoryMessage(ctx context.Context, msg *nats.Msg) error {
	ev, err := decodeHistoryMsg(msg)
	if err != nil {
		return fmt.Errorf("decode history msg: %w", err)
	}

	subj := intNats.StreamInstanceStatus.Subject(ev.Namespace, ev.InstanceID.String())
	pubID := "instance::status::" + ev.Namespace + "::" + ev.InstanceID.String() + "::" + strconv.FormatUint(ev.Sequence, 10)

	for attempt := range 10 {
		// 1) Read the last status for this order (if any).
		st, err := p.getLastStatusForSubject(ctx, subj)
		if err != nil {
			return err
		}
		if st == nil {
			st = &engine.InstanceStatus{}
		}
		// 2) If our event is not newer, we’re done (idempotent / monotonic).
		if ev.Sequence <= st.HistorySequence {
			return nil
		}
		// 3) Build payload
		applyEventToStatus(st, ev)
		body, _ := json.Marshal(st)
		msg := &nats.Msg{
			Subject: subj,
			Header:  nats.Header{},
			Data:    body,
		}
		// 4) Publish with dedupe + optimistic concurrency
		opts := []nats.PubOpt{
			nats.MsgId(pubID),
			nats.ExpectStream(intNats.StreamInstanceStatus.String()),
			nats.ExpectLastSequencePerSubject(st.Sequence),
		}
		_, err = p.js.PublishMsg(msg, opts...)
		if err == nil {
			// Update cache immediately to keep endpoint fresh.
			// p.cache.Upsert(*st)
			return nil
		}
		// If conflict, loop to re-read and retry.
		if isConcurrencyConflict(err) {
			continue
		}
		// If dedupe window hit (Msg-Id already applied), treat as success.
		if isDuplicate(err) {
			return nil
		}
		// Other errors: transient retry
		slog.Error("publish status msg", "err", err, "attempt", attempt)
		time.Sleep(50 * time.Millisecond)
	}

	return errors.New("failed to upsert status after retries")
}

func (p *projector) getLastStatusForSubject(ctx context.Context, subject string) (st *engine.InstanceStatus, err error) {
	msg, err := p.js.GetLastMsg(
		intNats.StreamInstanceStatus.String(),
		subject, nats.Context(ctx))
	if err != nil && errors.Is(err, nats.ErrMsgNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(msg.Data, &st)
	if err != nil {
		return nil, err
	}
	st.Sequence = msg.Sequence

	return st, nil
}

func decodeHistoryMsg(msg *nats.Msg) (*engine.InstanceEvent, error) {
	var ev engine.InstanceEvent
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

func applyEventToStatus(st *engine.InstanceStatus, ev *engine.InstanceEvent) {
	st.Status = ev.Type
	st.HistorySequence = ev.Sequence //

	switch ev.Type {
	case "pending":
		st.InstanceID = ev.InstanceID
		st.Namespace = ev.Namespace
		st.Metadata = ev.Metadata
		st.Script = ev.Script
		st.Mappings = ev.Mappings
		st.Fn = ev.Fn
		st.Input = ev.Input
		st.CreatedAt = ev.Time
	case "started":
		st.StartedAt = ev.Time
	case "failed":
		st.EndedAt = ev.Time
		st.Memory = ev.Memory
		st.Output = ev.Output
		st.Error = ev.Error
	case "succeeded":
		st.EndedAt = ev.Time
		st.Memory = ev.Memory
		st.Output = ev.Output
		st.Error = ev.Error
	}
}

func isConcurrencyConflict(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()

	return strings.Contains(msg, "wrong last sequence") ||
		strings.Contains(msg, "wrong last sequence per subject")
}

func isDuplicate(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())

	return strings.Contains(msg, "duplicate") ||
		strings.Contains(msg, "wrong last msg id")
}
