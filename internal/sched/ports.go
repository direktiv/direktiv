package sched

import (
	"time"

	"github.com/nats-io/nats.go"
)

type JetStream interface {
	Publish(subj string, data []byte, opts ...nats.PubOpt) (*nats.PubAck, error)
	Subscribe(subj string, cb nats.MsgHandler, opts ...nats.SubOpt) (*nats.Subscription, error)
}

type RuleStore interface {
	Snapshot(ns string) []*Rule
	Upsert(rule *Rule)
}

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }
