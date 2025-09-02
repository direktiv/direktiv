package pubsub

import (
	"log/slog"

	"github.com/direktiv/direktiv/internal/core"
	"github.com/nats-io/nats.go"
)

type PubSub struct {
	nc *nats.Conn
}

func NewPubSub(conn *nats.Conn) core.PubSub {
	return &PubSub{
		nc: conn,
	}
}

func (b *PubSub) Loop(circuit *core.Circuit) error {
	<-circuit.Done()
	return b.nc.Drain()
}

func (b *PubSub) Subscribe(channel core.Type, handler func(data []byte)) {
	_, err := b.nc.Subscribe(string(channel), func(msg *nats.Msg) {
		slog.Debug("received message", slog.String("channel", msg.Subject))
		handler(msg.Data)
	})
	if err != nil {
		// we can not recover here
		panic("can not subscribe to channel")
	}
}

func (b *PubSub) Publish(channel core.Type, data []byte) error {
	if data == nil {
		data = []byte("")
	}

	return b.nc.Publish(string(channel), data)
}
