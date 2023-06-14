package entwrapper

import (
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/ent"
)

func (db *Database) entNamespace(ns *ent.Namespace) *database.Namespace {
	if ns == nil {
		return nil
	}

	return &database.Namespace{
		ID:        ns.ID,
		CreatedAt: ns.CreatedAt,
		UpdatedAt: ns.UpdatedAt,
		Config:    ns.Config,
		Name:      ns.Name,
	}
}
