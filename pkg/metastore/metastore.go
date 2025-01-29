package metastore

type Store interface {
	EventsStore() EventsStore
	LogStore() LogStore
	TimelineStore() TimelineStore
}
