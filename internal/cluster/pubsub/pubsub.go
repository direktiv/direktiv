package pubsub

type Subject string

const (
	SubjFileSystemChange Subject = "filesystem.change"
	SubjNamespacesChange Subject = "namespace.change"
	SubjCacheDelete      Subject = "cache.delete"
	SubjServiceIgnite    Subject = "service.ignite"
	SubjRuntimeVariableSet Subject = "runtimevariable.set" 
)

type Handler func(data []byte)

type EventBus interface {
	Publish(subject Subject, data []byte) error
	Subscribe(subject Subject, h Handler) error

	Close() error
}
