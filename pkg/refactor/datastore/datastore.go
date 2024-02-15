package datastore

import (
	"errors"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/events"
	"github.com/direktiv/direktiv/pkg/refactor/logengine"
)

// Direktiv application data (namespaces, mirrors, etc..) are stored in a sql database. For each different
// application data, there is a MirrorStore responsible for doing all the reading and writing database operations.

// Store object wraps all different direktiv application stores.
//
//nolint:interfacebloat
type Store interface {
	Namespaces() core.NamespacesStore

	// Mirror returns datastore.MirrorStore, is responsible for reading and writing mirrors information.
	Mirror() MirrorStore

	// Logs returns logengine.LogStore, is responsible for reading and writing logs
	Logs() logengine.LogStore

	Secrets() SecretsStore

	RuntimeVariables() core.RuntimeVariablesStore

	EventHistory() events.EventHistoryStore
	EventListener() events.EventListenerStore
	EventListenerTopics() events.EventTopicsStore
	StagingEvents() events.StagingEventStore
}

// ErrNotFound is a common error type that should be returned by any store implementation
// for the error cases when getting a single entry failed due to none existence.
var ErrNotFound = errors.New("ErrNotFound")
