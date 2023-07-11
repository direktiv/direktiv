package cluster

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/nsqio/go-nsq"
)

// InstanceChannel wraps the provided string with some additional information,
// which when used as the channel argument to the Subscribe function, ensures that
// exactly one such channel per instance receives the message.
func (node *Node) InstanceChannel(channel string) string {
	return fmt.Sprintf("%s:%s", node.busChannelName, channel)
}

// UniqueChannel wraps the provided string with some additional information,
// which when used as the channel argument to the Subscribe function, ensures that
// this channel is unique so that all such channels receives the message even if
// multiple such channels exist on the same instance.
func (node *Node) UniqueChannel(channel string) string {
	return fmt.Sprintf("%s:%s", channel, uuid.New().String())
}

// Publish sends a message on a given topic to all relevant channels in the cluster.
func (node *Node) Publish(topic string, message []byte) error {
	return node.producer.Publish(topic, message)
}

// Subscribe registers a handler function to be called when a message is received on
// the given topic and channel. It is recommended to read the nsq docs about topics and
// channels to understand what values to use here.
//
// In short, a topic is something in common between all related publishers and subscribers.
// For each message published to a topic, each registered channel will receive exactly one
// copy of the message. If your channel is uniquely used by you, then you should get a copy
// of every message. If your channel is not unique, any single subscriber on that channel
// will receive a given message.
//
// To have a one-to-one publish-subscribe across the cluster, use a fixed constant. To
// ensure each node in the cluster gets a copy of each message, it is recommended to call
// InstanceChannel() to wrap your channel name. If you need your subscribe to get every
// message even if duplicate subscribers may exist on the same instance, use UniqueChannel().
//
// This function returns a function pointer than can be used to unsubscribe.
func (node *Node) Subscribe(topic, channel string, handler func(m []byte) error) (func(), error) {
	consumer, err := node.doSubscribe(topic, channel, handler)
	if err != nil {
		return nil, err
	}

	return consumer.consumer.Stop, nil
}

type messageConsumer struct {
	topic    string
	executor func(m []byte) error
	consumer *nsq.Consumer
}

func (h *messageConsumer) HandleMessage(m *nsq.Message) error {
	return h.executor(m.Body)
}
