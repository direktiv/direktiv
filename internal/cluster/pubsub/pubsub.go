package pubsub

type Subject string

const (
	SubjFileSystemChange Subject = "filesystem.change"
	SubjNamespacesChange Subject = "namespace.change"
	SubjCacheDelete      Subject = "cache.delete"
)

type Handler func(data []byte)

type EventBus interface {
	Publish(subject Subject, data []byte) error
	Subscribe(subject Subject, h Handler) error

	Flush() error
}
