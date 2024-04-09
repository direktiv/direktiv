package datastore

import (
	"errors"

	"github.com/direktiv/direktiv/pkg/refactor/events"
)

// Direktiv application data (namespaces, mirrors, etc..) are stored in a sql database. For each different
// application data, there is a MirrorStore responsible for doing all the reading and writing database operations.

// Store object wraps all different direktiv application stores.
//
//nolint:interfacebloat
type Store interface {
	Namespaces() NamespacesStore

	// Mirror returns datastore.MirrorStore, is responsible for reading and writing mirrors information.
	Mirror() MirrorStore

	NewLogs() LogStore

	Secrets() SecretsStore

	RuntimeVariables() RuntimeVariablesStore

	EventHistory() events.EventHistoryStore
	EventListener() events.EventListenerStore
	EventListenerTopics() events.EventTopicsStore
	StagingEvents() events.StagingEventStore
}

var (
	// ErrNotFound is a common error type that should be returned by any store implementation
	// for the error cases when getting a single entry failed due to none existence.
	ErrNotFound = errors.New("ErrNotFound")

	// ErrDuplication is a common error type that should be returned by any store implementation
	// when tying to violate unique constraints.
	ErrDuplication = errors.New("ErrDuplication")
)
