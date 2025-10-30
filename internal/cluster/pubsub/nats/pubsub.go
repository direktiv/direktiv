package nats

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/direktiv/direktiv/internal/cluster/pubsub"
	"github.com/nats-io/nats.go"
)

type Bus struct {
	nc     *nats.Conn
	logger *slog.Logger
	js     nats.JetStreamContext
}

func (b *Bus) Close() error {
	return b.nc.Drain()
}

type NatsConnect func() (*nats.Conn, error)

func New(nc NatsConnect, logger *slog.Logger) (*Bus, error) {
	if logger != nil {
		logger = logger.With("component", "cluster-pubsub")
	} else {
		logger = slog.New(slog.DiscardHandler)
	}

	conn, err := nc()
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}

	js, err := conn.JetStream()
	if err != nil {
		return nil, fmt.Errorf("creating jetstream: %w", err)
	}

	return &Bus{nc: conn, logger: logger, js: js}, nil
}

func (b *Bus) Subscribe(ctx context.Context, subject string, h pubsub.Handler) error {
	_, err := b.js.Subscribe(subject, func(msg *nats.Msg) {
		h(msg.Data)

	}, nats.AckNone())

	return err
}

func (b *Bus) Publish(ctx context.Context, subject string, data []byte) error {
	_, err := b.js.Publish(subject, data, nats.Context(ctx))

	return err
}
