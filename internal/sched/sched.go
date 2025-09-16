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
)

type Scheduler struct {
	js    JetStream
	cache *RuleCache
	clk   Clock
}

func New(js nats.JetStreamContext) *Scheduler {
	return &Scheduler{js: js, cache: NewRulesCache(), clk: realClock{}}
}

func (s *Scheduler) Start(lc *lifecycle.Manager) error {
	err := s.startRuleSubscription(lc.Context())
	if err != nil {
		return fmt.Errorf("start status cache: %w", err)
	}

	startTicking(lc, time.Second, s.processDueRules)

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
	id := fmt.Sprintf("%s-%d", rule.ID, runAt.UTC().Unix())
	data, _ := json.Marshal(Task{
		ID:           id,
		Namespace:    rule.Namespace,
		WorkflowPath: rule.WorkflowPath,
		RunAt:        s.clk.Now(),
		CreatedAt:    s.clk.Now(),
	})
	subject := fmt.Sprintf(intNats.SubjSchedTask, rule.Namespace, rule.ID)
	_, err := s.js.Publish(subject, data,
		nats.ExpectStream(intNats.StreamSchedTask),
		// important to ensure dedupe. we don't want to publish the same task twice from two different servers
		nats.MsgId(fmt.Sprintf("sched::task::%s", id)),
	)
	if err != nil {
		return fmt.Errorf("nats publish task, subj: %s, err: %w", subject, err)
	}

	// advance rule, persist
	next := runAt.Add(time.Duration(rule.CronExpr) * time.Second)
	rule.RunAt = next
	rule.UpdatedAt = s.clk.Now()

	// optimistic update in rule stream
	data, _ = json.Marshal(rule)
	subject = fmt.Sprintf(intNats.SubjSchedRule, rule.Namespace, rule.ID)
	_, err = s.js.Publish(subject, data,
		nats.ExpectStream(intNats.StreamSchedRule),
		nats.ExpectLastSequence(rule.Sequence),
		nats.MsgId(fmt.Sprintf("sched::rule::%s", rule.Fingerprint())),
	)
	if err != nil {
		return fmt.Errorf("nats publish rule update, subj: %s, err: %w", subject, err)
	}

	s.cache.Upsert(rule)

	return nil
}

func (s *Scheduler) processDueRules() error {
	// snapshot rules in cache
	rules := s.cache.Snapshot("")
	for _, rule := range rules {
		err := s.dispatchIfDue(rule)
		if err != nil {
			slog.Error("dispatchIfDue", "err", err, "id", rule.ID)
		} else {
			slog.Info("dispatchIfDue", "id", rule.ID)
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

	subject := fmt.Sprintf(intNats.SubjSchedRule, rule.Namespace, rule.ID)

	_, err = s.js.Publish(subject, data,
		nats.Context(ctx),
		nats.ExpectStream(intNats.StreamSchedRule),
		nats.MsgId(fmt.Sprintf("sched::rule::%s", rule.Fingerprint())),
	)

	return rule, err
}

func (s *Scheduler) ListRules(ctx context.Context) ([]*Rule, error) {
	data := s.cache.Snapshot("")

	return data, nil
}

func (s *Scheduler) startRuleSubscription(ctx context.Context) error {
	subj := fmt.Sprintf(intNats.SubjSchedRule, "*", "*")
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
		s.cache.Upsert(&rule)
	}, nats.AckNone())
	if err != nil {
		return err
	}

	return nil
}
