package nats

import (
	"fmt"
	"log/slog"

	"github.com/direktiv/direktiv/internal/cluster/pubsub"
	"github.com/nats-io/nats.go"
)

type Bus struct {
	nc     *nats.Conn
	logger *slog.Logger
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

	return &Bus{nc: conn, logger: logger}, nil
}

func (b *Bus) Publish(subject pubsub.Subject, data []byte) error {
	err := b.nc.Publish(string(subject), data)
	if err != nil {
		return err
	}

	return nil
}

func (b *Bus) Subscribe(subject pubsub.Subject, h pubsub.Handler) error {
	wrapper := func(msg *nats.Msg) {
		// Protect handlers; never let a panic kill the NATS dispatcher.
		defer func() {
			if r := recover(); r != nil {
				b.logger.Error("panic in pubsub handler", "subject", msg.Subject, "recover", r)
			}
		}()

		h(msg.Data)
	}

	_, err := b.nc.Subscribe(string(subject), wrapper)
	if err != nil {
		return err
	}

	return nil
}
