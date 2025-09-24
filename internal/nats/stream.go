package nats

import (
	"fmt"
	"strings"

	"github.com/nats-io/nats.go"
)

type StreamDescriptor string

func (n StreamDescriptor) Subject(namespace string, id string) string {
	return string(n) + fmt.Sprintf(".%s.%s", namespace, id)
}

func (n StreamDescriptor) String() string {
	return strings.ReplaceAll(string(n), ".", "-")
}

func (n StreamDescriptor) PullSubscribe(js nats.JetStreamContext, opts ...nats.SubOpt) (*nats.Subscription, error) {
	opts = append(opts, nats.BindStream(n.String()))
	sub, err := js.PullSubscribe(n.Subject("*", "*"), n.String(), opts...)
	if err != nil {
		return nil, err
	}

	return sub, nil
}
