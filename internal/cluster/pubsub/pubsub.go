package pubsub

import (
	"context"
)

type Subject string

type Handler func(data []byte)

type EventBus interface {
	Publish(ctx context.Context, subject string, data []byte) error
	Subscribe(ctx context.Context, subject string, h Handler) error

	Close() error
}
