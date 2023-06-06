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

func (db *Database) entAnnotation(annotation *ent.Annotation) *database.Annotation {
	if annotation == nil {
		return nil
	}

	return &database.Annotation{
		ID:        annotation.ID,
		Name:      annotation.Name,
		CreatedAt: annotation.CreatedAt,
		UpdatedAt: annotation.UpdatedAt,
		Size:      annotation.Size,
		Hash:      annotation.Hash,
		Data:      annotation.Data,
		MimeType:  annotation.MimeType,
	}
}

func (db *Database) entVarRef(vref *ent.VarRef) *database.VarRef {
	if vref == nil {
		return nil
	}

	return &database.VarRef{
		ID:        vref.ID,
		Name:      vref.Name,
		Behaviour: vref.Behaviour,
		VarData:   vref.Edges.Vardata.ID,
	}
}

func (db *Database) entVarData(v *ent.VarData) *database.VarData {
	if v == nil {
		return nil
	}

	return &database.VarData{
		ID:        v.ID,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
		Size:      v.Size,
		Hash:      v.Hash,
		Data:      v.Data,
		MimeType:  v.MimeType,
	}
}
