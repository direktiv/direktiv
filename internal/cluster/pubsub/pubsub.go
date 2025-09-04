package pubsub

import "github.com/direktiv/direktiv/internal/core"

type Subject string

const (
	SubjFileSystemChange Subject = "filesystem.change"
	SubjNamespacesChange Subject = "namespace.change"
	SubjCacheDelete      Subject = "cache.delete"
)

type Bus interface {
	Subscribe(channel Subject, handler func(data []byte))
	Publish(channel Subject, data []byte) error
	Loop(circuit *core.Circuit) error
}
