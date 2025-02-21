package metastore

type Store interface {
	// EventsStore() EventsStore
	LogStore() LogStore
	// MetricsStore() MetricsStore
	// TimelineStore() TimelineStore
}
