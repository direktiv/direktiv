package cache

type Cache interface {
	SetTTL(key string, value any, ttl int)
	Set(key string, value any)
	Delete(key string)
	Get(key string) (any, bool)

	Hits() uint64
	Misses() uint64

	Close()
}
