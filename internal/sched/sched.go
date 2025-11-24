package sched

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/direktiv/direktiv/internal/engine"
	intNats "github.com/direktiv/direktiv/internal/nats"
	"github.com/direktiv/direktiv/pkg/lifecycle"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"k8s.io/utils/clock"
)

type Scheduler struct {
	js         JetStream
	cache      *RuleCache
	clk        clock.WithTicker
	lg         *slog.Logger
	engine     *engine.Engine
	withEngine bool
}

func New(js JetStream, engine *engine.Engine, clk clock.WithTicker, lg *slog.Logger) *Scheduler {
	return &Scheduler{js: js, engine: engine, withEngine: true, cache: NewRulesCache(), clk: clk, lg: lg}
}

func NewWithoutEngine(js JetStream, clk clock.WithTicker, lg *slog.Logger) *Scheduler {
	return &Scheduler{js: js, withEngine: false, cache: NewRulesCache(), clk: clk, lg: lg}
}

func (s *Scheduler) Start(lc *lifecycle.Manager) error {
	err := s.startRuleSubscription(lc.Context())
	if err != nil {
		return fmt.Errorf("start status cache: %w", err)
	}
	if s.withEngine {
		err = s.startTaskSubscription(lc.Context())
		if err != nil {
			return fmt.Errorf("start task worker: %w", err)
		}
	}

	startTicking(lc, s.clk, 100*time.Millisecond, s.processDueRules)

	return nil
}

func (s *Scheduler) dispatchIfDue(rule *Rule) error {
	rule = rule.Clone()
	now := s.clk.Now()
	runAt := rule.RunAt

	if rule.RunAt.IsZero() || runAt.After(now) {
		return fmt.Errorf("not-due")
	}

	// publish task
	id := rule.ID + "-" + runAt.Format("20060102150405")
	data, _ := json.Marshal(Task{
		ID:           id,
		Namespace:    rule.Namespace,
		WorkflowPath: rule.WorkflowPath,
		RunAt:        s.clk.Now(),
		CreatedAt:    s.clk.Now(),
	})
	subject := intNats.StreamSchedTask.Subject(rule.Namespace, rule.ID)
	_, err := s.js.Publish(subject, data,
		nats.ExpectStream(intNats.StreamSchedTask.String()),
		// important to ensure dedupe. we don't want to publish the same task twice from two different servers
		nats.MsgId(fmt.Sprintf("sched::task::%s", id)),
	)
	if err != nil {
		return fmt.Errorf("nats publish task, subj: %s, err: %w", subject, err)
	}
	s.lg.Debug("published task", "msgID", id)

	// advance rule, persist
	next, err := CalculateCronExpr(rule.CronExpr, runAt)
	if err != nil {
		return fmt.Errorf("calculate next run: %w", err)
	}
	rule.RunAt = next
	rule.UpdatedAt = s.clk.Now()

	// optimistic update in rule stream
	data, _ = json.Marshal(rule)
	subject = intNats.StreamSchedRule.Subject(rule.Namespace, rule.ID)
	_, err = s.js.Publish(subject, data,
		nats.ExpectStream(intNats.StreamSchedRule.String()),
		nats.ExpectLastSequencePerSubject(rule.Sequence),
		nats.MsgId(fmt.Sprintf("sched::rule::%s", rule.Fingerprint())),
	)
	if err != nil {
		return fmt.Errorf("nats publish rule update, err: %w", err)
	}

	return nil
}

func (s *Scheduler) processDueRules() error {
	// snapshot rules in cache
	rules := s.cache.Snapshot("")
	for _, rule := range rules {
		err := s.dispatchIfDue(rule)
		if err != nil && err.Error() == "not-due" {
			continue
		}
		if err != nil {
			s.lg.Error("dispatchIfDue", "err", err, "id", rule.ID)
		} else {
			s.lg.Debug("dispatchIfDue", "id", rule.ID)
		}
	}

	return nil
}

func (s *Scheduler) SetRule(ctx context.Context, rule *Rule) (*Rule, error) {
	// clone to protect against mutation
	if rule != nil {
		clone := *rule // value copy (shallow)
		rule = &clone  // rule now points to the new copy
	} else {
		return nil, fmt.Errorf("nil rule")
	}

	rule.ID = CalculateRuleID(*rule)
	rule.CreatedAt = s.clk.Now()
	rule.UpdatedAt = s.clk.Now()

	data, err := json.Marshal(rule)
	if err != nil {
		return nil, fmt.Errorf("marshal rule: %w", err)
	}

	subject := intNats.StreamSchedRule.Subject(rule.Namespace, rule.ID)

	_, err = s.js.Publish(subject, data,
		nats.Context(ctx),
		nats.ExpectStream(intNats.StreamSchedRule.String()),
		nats.MsgId(fmt.Sprintf("sched::rule::%s", rule.Fingerprint())),
	)

	return rule, err
}

func (s *Scheduler) ListRules(ctx context.Context) ([]*Rule, error) {
	data := s.cache.Snapshot("")

	return data, nil
}

func (s *Scheduler) startRuleSubscription(ctx context.Context) error {
	subj := intNats.StreamSchedRule.Subject("*", "*")
	// ephemeral, AckNone (we don't want to disturb the stream/consumers)
	_, err := s.js.Subscribe(subj, func(msg *nats.Msg) {
		var rule Rule
		if err := json.Unmarshal(msg.Data, &rule); err != nil {
			// best-effort; ignore bad payloads
			return
		}
		meta, err := msg.Metadata()
		if err != nil {
			// best-effort; ignore bad payloads
			return
		}
		rule.Sequence = meta.Sequence.Stream
		s.lg.Debug("rule upsert from steam", "id", rule.ID)
		s.cache.Upsert(&rule)
	}, nats.AckNone())
	if err != nil {
		return err
	}

	return nil
}

func (s *Scheduler) startTaskSubscription(ctx context.Context) error {
	subj := intNats.StreamSchedTask.Subject("*", "*")
	_, err := s.js.Subscribe(subj, func(msg *nats.Msg) {
		// Immediately bail out if the context is cancelled
		if ctx.Err() != nil {
			_ = msg.Nak() // best-effort
			return
		}

		var tsk Task
		if err := json.Unmarshal(msg.Data, &tsk); err != nil {
			// best-effort; ignore bad payloads
			return
		}
		s.lg.Info("task received from stream", "id", tsk.ID, "ns", tsk.Namespace, "wf", tsk.WorkflowPath)
		//nolint:contextcheck
		_, _, err := s.engine.StartWorkflow(context.Background(), uuid.New(), tsk.Namespace, tsk.WorkflowPath, "{}", map[string]string{
			engine.LabelWithNotify:   strconv.FormatBool(false),
			engine.LabelWithSyncExec: strconv.FormatBool(false),
			engine.LabelInvokerType:  "cron",
			engine.LabelWithScope:    "main",
		})
		if err != nil {
			s.lg.Error("failed to start cron workflow", "err", err, "id", tsk.ID, "ns", tsk.Namespace, "wf", tsk.WorkflowPath)
			if ackErr := msg.Nak(); ackErr != nil {
				s.lg.Error("failed to nak task message", "err", ackErr, "id", tsk.ID)
			}

			return
		}
		s.lg.Info("started cron workflow", "id", tsk.ID, "ns", tsk.Namespace, "wf", tsk.WorkflowPath)
		if err := msg.Ack(); err != nil {
			s.lg.Error("failed to ack task message", "err", err, "id", tsk.ID)
		}
	}, nats.ManualAck(),
		nats.AckExplicit(), // Explicit ack policy
		nats.MaxAckPending(1024),
		nats.Context(ctx))
	if err != nil {
		return fmt.Errorf("nats subscribe task: %w", err)
	}

	return nil
}
