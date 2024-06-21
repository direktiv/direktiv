package tsengine

import (
	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/database"
)

func NewManager(db *database.SQLStore) Manager {
	return Manager{
		db: db,
	}
}

type Manager struct {
	db *database.SQLStore
}

func (tsm Manager) Run(circuit *core.Circuit) error {
	return nil
}
