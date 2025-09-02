package core

type Cache interface {
	SetTTL(key string, value any, ttl int)
	Set(key string, value any)
	Delete(key string)
	Get(key string) (any, bool)
	Run(circuit *Circuit)

	Hits() uint64
	Misses() uint64
}
