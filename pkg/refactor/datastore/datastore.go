package datastore

import (
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/mirror"
)

type Store interface {
	Mirror() mirror.Store
	FileAttributes() core.FileAttributesStore
}
