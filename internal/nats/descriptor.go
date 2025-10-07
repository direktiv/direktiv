package nats

import (
	"fmt"
	"strings"

	"github.com/nats-io/nats.go"
)

type Descriptor struct {
	name           string
	streamConfig   *nats.StreamConfig
	consumerConfig *nats.ConsumerConfig
}

func newDescriptor(name string, streamConfig *nats.StreamConfig, consumerConfig *nats.ConsumerConfig) *Descriptor {
	desc := &Descriptor{
		name: name,
	}

	streamConfig.Name = desc.String()
	streamConfig.Subjects = []string{
		desc.Subject("*", "*"),
	}

	if consumerConfig != nil {
		consumerConfig.Durable = desc.String()
		consumerConfig.FilterSubject = desc.Subject("*", "*")
	}

	desc.streamConfig = streamConfig
	desc.consumerConfig = consumerConfig

	return desc
}

func (n Descriptor) Subject(namespace string, id string) string {
	// replace dots with dashes as NATS does not allow dots in subjects.
	namespace = strings.ReplaceAll(namespace, ".", "-")
	return n.name + fmt.Sprintf(".%s.%s", namespace, id)
}

func (n Descriptor) String() string {
	return strings.ReplaceAll(n.name, ".", "-")
}

func (n Descriptor) PullSubscribe(js nats.JetStreamContext, opts ...nats.SubOpt) (*nats.Subscription, error) {
	opts = append(opts, nats.BindStream(n.String()))
	sub, err := js.PullSubscribe(n.Subject("*", "*"), n.String(), opts...)
	if err != nil {
		return nil, err
	}

	return sub, nil
}
