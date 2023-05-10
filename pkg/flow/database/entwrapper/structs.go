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

// TODO: delete this.
func EntInstance(inst *ent.Instance) *database.Instance {
	return entInstance(inst)
}

func entInstance(inst *ent.Instance) *database.Instance {
	if inst == nil {
		return nil
	}

	x := &database.Instance{
		ID:           inst.ID,
		CreatedAt:    inst.CreatedAt,
		UpdatedAt:    inst.UpdatedAt,
		EndAt:        inst.EndAt,
		Status:       inst.Status,
		As:           inst.As,
		ErrorCode:    inst.ErrorCode,
		ErrorMessage: inst.ErrorMessage,
		Invoker:      inst.Invoker,
		CallPath:     inst.Callpath,
		Revision:     *inst.RevisionID,
		Workflow:     *inst.WorkflowID,
	}

	if inst.Edges.Namespace != nil {
		x.Namespace = inst.Edges.Namespace.ID
	}

	if inst.Edges.Runtime != nil {
		x.Runtime = inst.Edges.Runtime.ID
	}

	return x
}

// TODO: delete this.
func EntInstanceRuntime(rt *ent.InstanceRuntime) *database.InstanceRuntime {
	return entInstanceRuntime(rt)
}

func entInstanceRuntime(rt *ent.InstanceRuntime) *database.InstanceRuntime {
	if rt == nil {
		return nil
	}

	x := &database.InstanceRuntime{
		ID:              rt.ID,
		Input:           rt.Input,
		Data:            rt.Data,
		Controller:      rt.Controller,
		Memory:          rt.Memory,
		Flow:            rt.Flow,
		Output:          rt.Output,
		StateBeginTime:  rt.StateBeginTime,
		Deadline:        rt.Deadline,
		Attempts:        rt.Attempts,
		CallerData:      rt.CallerData,
		InstanceContext: rt.InstanceContext,
		StateContext:    rt.StateContext,
		Metadata:        rt.Metadata,
		LogToEvents:     rt.LogToEvents,
	}

	if rt.Edges.Caller != nil {
		x.Caller = rt.Edges.Caller.ID
	}

	return x
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
