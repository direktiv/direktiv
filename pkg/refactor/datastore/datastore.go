package datastore

import (
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/logengine"
	"github.com/direktiv/direktiv/pkg/refactor/mirror"
)

// Direktiv application data (namespaces, annotations, mirrors, etc..) are stored in a sql database. For each different
// application data, there is a Store responsible for doing all the reading and writing database operations.

// Store object wraps all different direktiv application stores.
type Store interface {
	// Mirror returns mirror.Store, is responsible for reading and writing mirrors information.
	Mirror() mirror.Store
	// FileAnnotations returns core.FileAnnotationsStore, is responsible for reading and writing file annotations
	// information.
	FileAnnotations() core.FileAnnotationsStore
	// Logs returns logengine.LogStore, is responsible for reading and writing logs
	Logs() logengine.LogStore

	Secrets() core.SecretsStore
}
