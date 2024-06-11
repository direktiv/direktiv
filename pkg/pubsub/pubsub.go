package pubsub

type CoreBus interface {
	Publish(channel string, data string) error
	Listen(channel string) error
	Loop(done <-chan struct{}, handler func(channel string, data string)) error
}
