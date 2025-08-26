package pubsub

import (
	"log/slog"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/nats-io/nats.go"
)

type Type string

const (
	FileSystemChangeEvent Type = "filesystem.change"
	NamespacesChangeEvent Type = "namespace.change"
	CacheDeleteEvent      Type = "cache.delete"
)

type Bus struct {
	nc *nats.Conn
}

func NewBus(conn *nats.Conn) *Bus {
	return &Bus{
		nc: conn,
	}
}

func (b *Bus) Loop(circuit *core.Circuit) error {
	<-circuit.Done()
	return b.nc.Drain()
}

func (b *Bus) Subscribe(channel Type, handler func(data []byte)) {
	_, err := b.nc.Subscribe(string(channel), func(msg *nats.Msg) {
		slog.Debug("received message", slog.String("channel", msg.Subject))
		handler(msg.Data)
	})
	if err != nil {
		// we can not recover here
		panic("can not subscribe to channel")
	}
}

func (b *Bus) Publish(channel Type, data []byte) error {
	if data == nil {
		data = []byte("")
	}

	return b.nc.Publish(string(channel), data)
}
