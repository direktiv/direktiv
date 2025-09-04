package pubsub

import "github.com/direktiv/direktiv/internal/core"

type Type string

const (
	FileSystemChangeEvent Type = "filesystem.change"
	NamespacesChangeEvent Type = "namespace.change"
	CacheDeleteEvent      Type = "cache.delete"
)

type Bus interface {
	Subscribe(channel Type, handler func(data []byte))
	Publish(channel Type, data []byte) error
	Loop(circuit *core.Circuit) error
}
