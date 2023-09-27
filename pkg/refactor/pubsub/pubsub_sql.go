package pubsub

type Bus interface {
	Publish(channel string, data string) error
	Subscribe(channel string, handler func(data string))
}
