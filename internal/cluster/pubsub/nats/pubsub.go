package nats

import (
	"io"
	"log/slog"

	"github.com/nats-io/nats.go"

	"github.com/direktiv/direktiv/internal/cluster/pubsub"
)

type Bus struct {
	nc     *nats.Conn
	logger *slog.Logger
}

func New(nc *nats.Conn, logger *slog.Logger) *Bus {
	if logger != nil {
		logger = logger.With("component", "cluster-pubsub")
	} else {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return &Bus{nc: nc, logger: logger}
}

func (b *Bus) Publish(subject pubsub.Subject, data []byte) error {
	err := b.nc.Publish(string(subject), data)
	if err != nil {
		return err
	}

	return nil
}

func (b *Bus) Flush() error {
	return b.nc.Flush()
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
