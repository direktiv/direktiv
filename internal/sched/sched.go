package sched

import (
	"context"
	"crypto/sha256"
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

type Config struct {
	ID           string `json:"id"`
	Namespace    string `json:"namespace"`
	WorkflowPath string `json:"workflowPath"`
	Kind         Kind   `json:"kind"`
	Status       Status `json:"status"`
	CronExpr     string `json:"cronExpr,omitempty"`

	RunAt     time.Time `json:"runAt"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Sequence uint64 `json:"sequence"`
}

func CalculateConfigID(c Config) string {
	str := fmt.Sprintf("ns:%s-path:%s", c.Namespace, c.WorkflowPath)
	sh := sha256.Sum256([]byte(str))

	return fmt.Sprintf("%x", sh[:10])
}

type UpdateConfig struct {
	Status    *Status
	RunAt     *time.Time
	UpdatedAt time.Time
}

type Store interface {
	SetConfig(ctx context.Context, cfg *Config) (*Config, error)
	ListConfigs(ctx context.Context) ([]*Config, error)
}

type Scheduler struct {
	js    nats.JetStreamContext
	cache *ConfigCache
}

func New(js nats.JetStreamContext) *Scheduler {
	return &Scheduler{js: js, cache: NewConfigCache()}
}

func (s *Scheduler) Start(lc *lifecycle.Manager) error {
	err := s.startConfigCache(lc.Context())
	if err != nil {
		return fmt.Errorf("start status cache: %w", err)
	}

	return nil
}

func (s *Scheduler) SetConfig(ctx context.Context, cfg *Config) (*Config, error) {
	// clone to protect against mutation
	if cfg != nil {
		clone := *cfg // value copy (shallow)
		cfg = &clone  // cfg now points to the new copy
	} else {
		return nil, fmt.Errorf("nil config")
	}

	cfg.ID = CalculateConfigID(*cfg)
	cfg.CreatedAt = time.Now()
	cfg.UpdatedAt = cfg.CreatedAt

	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshal cofig: %w", err)
	}

	subject := fmt.Sprintf(intNats.SubjSchedConfig, cfg.Namespace, cfg.ID)

	_, err = s.js.Publish(subject, data,
		nats.Context(ctx),
		//nats.MsgId(fmt.Sprintf("schedConfig::%s::%s", cfg.Namespace, cfg.ID)),
	)

	return cfg, err
}

func (s *Scheduler) ListConfigs(ctx context.Context) ([]*Config, error) {
	data := s.cache.Snapshot("")

	out := make([]*Config, len(data))
	for i, d := range data {
		out[i] = &d
	}

	return out, nil
}

func (s *Scheduler) startConfigCache(ctx context.Context) error {
	subj := fmt.Sprintf(intNats.SubjSchedConfig, "*", "*")
	// ephemeral, AckNone (we don't want to disturb the stream/consumers)
	_, err := s.js.Subscribe(subj, func(msg *nats.Msg) {
		var cfg Config
		if err := json.Unmarshal(msg.Data, &cfg); err != nil {
			// best-effort; ignore bad payloads
			return
		}
		s.cache.Upsert(cfg)
		_ = msg.Term() // AckNone/Term to be explicit; no re-delivery desired
	}, nats.ManualAck(), nats.AckNone())
	if err != nil {
		return err
	}

	return nil
}
