package sched

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

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
	CronExpr     string `json:"cronExpr,omitempty"`

	RunAt     time.Time `json:"runAt"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	DeletedAt time.Time `json:"deletedAt"`

	Sequence uint64 `json:"sequence"`
}

func (c *Rule) Fingerprint() string {
	// make a shallow copy with volatile fields zeroed
	cp := *c
	cp.CreatedAt = time.Time{}
	cp.UpdatedAt = time.Time{}

	b, _ := json.Marshal(&cp)
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

type Task struct {
	ID           string    `json:"id"`
	Namespace    string    `json:"namespace"`
	WorkflowPath string    `json:"workflowPath"`
	RunAt        time.Time `json:"runAt"`
	CreatedAt    time.Time `json:"createdAt"`
}

type JetStream interface {
	Publish(subj string, data []byte, opts ...nats.PubOpt) (*nats.PubAck, error)
	Subscribe(subj string, cb nats.MsgHandler, opts ...nats.SubOpt) (*nats.Subscription, error)
}
