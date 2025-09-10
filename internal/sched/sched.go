package sched

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	intNats "github.com/direktiv/direktiv/internal/nats"
	"github.com/direktiv/direktiv/pkg/lifecycle"
	"github.com/nats-io/nats.go"
)

type Kind string

const (
	KindOneTime Kind = "oneTime"
	KindCron    Kind = "cron"
)

type Rule struct {
	ID           string `json:"id"`
	Namespace    string `json:"namespace"`
	WorkflowPath string `json:"workflowPath"`
	Kind         Kind   `json:"kind"`
	CronExpr     int    `json:"cronExpr,omitempty"`

	RunAt     time.Time `json:"runAt"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Sequence uint64 `json:"-"`
}

type Task struct {
	ID           string    `json:"id"`
	Namespace    string    `json:"namespace"`
	WorkflowPath string    `json:"workflowPath"`
	RunAt        time.Time `json:"runAt"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (c *Rule) Fingerprint() string {
	// make a shallow copy with volatile fields zeroed
	cp := c
	cp.CreatedAt = time.Time{}
	cp.UpdatedAt = time.Time{}

	b, _ := json.Marshal(cp)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:8])
}

// Clone returns a copy of the rule.
func (c *Rule) Clone() *Rule {
	cp := *c

	return &cp
}

func CalculateRuleID(c Rule) string {
	str := fmt.Sprintf("ns:%s-path:%s", c.Namespace, c.WorkflowPath)
	sh := sha256.Sum256([]byte(str))

	return fmt.Sprintf("%x", sh[:10])
}

type Store interface {
	SetRule(ctx context.Context, rule *Rule) (*Rule, error)
	ListRules(ctx context.Context) ([]*Rule, error)
}

type Scheduler struct {
	js    nats.JetStreamContext
	cache *RulesCache
}

func New(js nats.JetStreamContext) *Scheduler {
	return &Scheduler{js: js, cache: NewRulesCache()}
}

func (s *Scheduler) Start(lc *lifecycle.Manager) error {
	err := s.startRuleCache(lc.Context())
	if err != nil {
		return fmt.Errorf("start status cache: %w", err)
	}

	lc.Go(func() error {
		t := time.NewTicker(time.Second)
		defer t.Stop()
		for {
			select {
			case <-lc.Done():
				return nil
			case <-t.C:
				err := s.tick()
				if err != nil {
					return fmt.Errorf("scheduler tick, err: %w", err)
				}
			}
		}
	})

	return nil
}

func (s *Scheduler) tick() error {
	now := time.Now().UTC()
	// snapshot rules in cache
	rules := s.cache.Snapshot("")
	for _, rule := range rules {
		fmt.Print("\n\n\n\n")

		if rule.RunAt.IsZero() || rule.RunAt.UTC().After(now) {
			fmt.Printf("skipping rule: %s\n", rule.ID)
			continue
		}

		fmt.Printf("triggering rule: %s\n", rule.ID)

		runAt := rule.RunAt.UTC()
		id := rule.ID + "-" + runAt.UTC().Format(time.RFC3339Nano)
		data, _ := json.Marshal(Task{
			ID:           id,
			Namespace:    rule.Namespace,
			WorkflowPath: rule.WorkflowPath,
			RunAt:        runAt,
			CreatedAt:    time.Now(),
		})
		subject := fmt.Sprintf(intNats.SubjSchedTask, rule.Namespace, rule.ID)
		_, err := s.js.Publish(subject, data,
			nats.Context(context.Background()),
			nats.ExpectStream(intNats.StreamSchedTask),
			// important to ensure dedupe. we don't want to publish the same task twice from two different servers
			nats.MsgId(fmt.Sprintf("sched::task::%s", id)),
		)
		if err != nil {
			slog.Error("nats publish task", "err", err, "subject", subject, "id", id)
			continue // try again next tick
		}

		// compute next run time
		runAt = runAt.Add(time.Duration(rule.CronExpr) * time.Second)
		rule.RunAt = runAt
		rule.UpdatedAt = time.Now()
		s.cache.Upsert(rule)

		// update rule
		data, _ = json.Marshal(rule)
		subject = fmt.Sprintf(intNats.SubjSchedRule, rule.Namespace, rule.ID)
		_, err = s.js.Publish(subject, data,
			nats.Context(context.Background()),
			nats.ExpectStream(intNats.StreamSchedRule),
			nats.ExpectLastSequence(rule.Sequence),
			nats.MsgId(fmt.Sprintf("sched::rule::%s", rule.Fingerprint())),
		)
		if err != nil {
			slog.Error("nats publish rule update", "err", err, "subject", subject)
			continue
		}

		fmt.Printf("suuccesfully triggered rule: %s\n", rule.ID)
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
	rule.CreatedAt = time.Now()
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

func (s *Scheduler) startRuleCache(ctx context.Context) error {
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
