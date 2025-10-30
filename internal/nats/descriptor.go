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
	dp := &Descriptor{
		name: name,
	}

	streamConfig.Name = dp.String()
	streamConfig.Subjects = []string{
		dp.Subject("*", "*"),
		name,
	}

	if consumerConfig != nil {
		consumerConfig.Durable = dp.String()
		consumerConfig.FilterSubject = dp.Subject("*", "*")
	}

	dp.streamConfig = streamConfig
	dp.consumerConfig = consumerConfig

	return dp
}

func (n Descriptor) Subject(namespace string, id string) string {
	// replace dots with dashes as NATS does not allow dots in subjects.
	namespace = strings.ReplaceAll(namespace, ".", "-")
	return n.name + fmt.Sprintf(".%s.%s", namespace, id)
}

func (n Descriptor) String() string {
	return strings.ReplaceAll(n.name, ".", "-")
}

func (n Descriptor) Name() string {
	return n.name
}

func (n Descriptor) PullSubscribe(js nats.JetStreamContext, opts ...nats.SubOpt) (*nats.Subscription, error) {
	opts = append(opts, nats.BindStream(n.String()))
	sub, err := js.PullSubscribe(n.Subject("*", "*"), n.String(), opts...)
	if err != nil {
		return nil, err
	}

	return sub, nil
}
