package sched

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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

type Status string

const (
	StatusActive    Status = "active"
	StatusCancelled Status = "cancelled"
)

type Rule struct {
	ID           string `json:"id"`
	Namespace    string `json:"namespace"`
	WorkflowPath string `json:"workflowPath"`
	Kind         Kind   `json:"kind"`
	Status       Status `json:"status"`
	CronExpr     string `json:"cronExpr,omitempty"`

	RunAt     time.Time `json:"runAt"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Sequence uint64 `json:"-"`
}

func (c Rule) Fingerprint() string {
	// make a shallow copy with volatile fields zeroed
	cp := c
	cp.CreatedAt = time.Time{}
	cp.UpdatedAt = time.Time{}

	b, _ := json.Marshal(cp)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:8])
}

func CalculateRuleID(c Rule) string {
	str := fmt.Sprintf("ns:%s-path:%s", c.Namespace, c.WorkflowPath)
	sh := sha256.Sum256([]byte(str))

	return fmt.Sprintf("%x", sh[:10])
}

type UpdateRule struct {
	Status    *Status
	RunAt     *time.Time
	UpdatedAt time.Time
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
		nats.MsgId(fmt.Sprintf("sched::rule::%s", rule.Fingerprint())),
	)

	return rule, err
}

func (s *Scheduler) ListRules(ctx context.Context) ([]*Rule, error) {
	data := s.cache.Snapshot("")

	out := make([]*Rule, len(data))
	for i, d := range data {
		out[i] = &d
	}

	return out, nil
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
		s.cache.Upsert(rule)
		_ = msg.Term() // AckNone/Term to be explicit; no re-delivery desired
	}, nats.ManualAck(), nats.AckNone())
	if err != nil {
		return err
	}

	return nil
}
