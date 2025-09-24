package sched

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	intNats "github.com/direktiv/direktiv/internal/nats"
	"github.com/direktiv/direktiv/pkg/lifecycle"
	"github.com/nats-io/nats.go"
	"k8s.io/utils/clock"
)

type Scheduler struct {
	js    JetStream
	cache *RuleCache
	clk   clock.WithTicker
	lg    *slog.Logger
}

func New(js JetStream, clk clock.WithTicker, lg *slog.Logger) *Scheduler {
	return &Scheduler{js: js, cache: NewRulesCache(), clk: clk, lg: lg}
}

func (s *Scheduler) Start(lc *lifecycle.Manager) error {
	err := s.startRuleSubscription(lc.Context())
	if err != nil {
		return fmt.Errorf("start status cache: %w", err)
	}
	startTicking(lc, s.clk, 100*time.Millisecond, s.processDueRules)

	return nil
}

func (s *Scheduler) dispatchIfDue(rule *Rule) error {
	rule = rule.Clone()
	now := s.clk.Now().UTC()
	runAt := rule.RunAt.UTC()

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
	next, err := calculateCronExpr(rule.CronExpr, runAt)
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
	rule.UpdatedAt = rule.CreatedAt

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
