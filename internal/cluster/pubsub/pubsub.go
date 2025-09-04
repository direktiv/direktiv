package pubsub

import (
	"context"
)

type Subject string

const (
	SubjFileSystemChange Subject = "filesystem.change"
	SubjNamespacesChange Subject = "namespace.change"
	SubjCacheDelete      Subject = "cache.delete"
)

type Message struct {
	Subject Subject
	Data    []byte
}

type Handler func(ctx context.Context, m Message) error

type EventBus interface {
	Subscribe(channel Subject, handler func(data []byte))
	Publish(channel Subject, data []byte) error
}
