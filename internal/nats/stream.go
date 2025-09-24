package nats

import "fmt"

type StreamDescriptor string

func (n StreamDescriptor) Subject(namespace string, id string) string {
	return string(n) + fmt.Sprintf(".%s.%s", namespace, id)
}

func (n StreamDescriptor) String() string {
	return string(n)
}
