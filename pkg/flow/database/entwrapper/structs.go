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
		Root:      ns.Edges.Inodes[0].ID,
	}
}

func entInode(ino *ent.Inode) *database.Inode {
	if ino == nil {
		return nil
	}

	var children []*database.Inode
	for _, x := range ino.Edges.Children {
		children = append(children, &database.Inode{
			ID:   x.ID,
			Name: x.Name,
		})
	}

	x := &database.Inode{
		ID:           ino.ID,
		CreatedAt:    ino.CreatedAt,
		UpdatedAt:    ino.UpdatedAt,
		Name:         ino.Name,
		Type:         ino.Type,
		Attributes:   ino.Attributes,
		ExtendedType: ino.ExtendedType,
		ReadOnly:     ino.ReadOnly,
		Children:     children,
		Namespace:    ino.Edges.Namespace.ID,
	}

	if ino.Edges.Parent == nil {
		if x.Name != "" {
			panic("failed to resolve inode parent")
		}
	} else {
		x.Parent = ino.Edges.Parent.ID
	}

	if ino.Edges.Workflow != nil {
		x.Workflow = ino.Edges.Workflow.ID
	}

	return x
}

func entWorkflow(wf *ent.Workflow) *database.Workflow {
	if wf == nil {
		return nil
	}

	var refs []*database.Ref
	for _, x := range wf.Edges.Refs {
		refs = append(refs, entRef(x))
	}

	var revisions []*database.Revision
	for _, x := range wf.Edges.Revisions {
		revisions = append(revisions, &database.Revision{
			ID:   x.ID,
			Hash: x.Hash,
		})
	}

	var routes []*database.Route
	for _, x := range wf.Edges.Routes {
		routes = append(routes, &database.Route{
			ID:     x.ID,
			Weight: x.Weight,
			Ref:    entRef(x.Edges.Ref),
		})
	}

	return &database.Workflow{
		ID:          wf.ID,
		Live:        wf.Live,
		LogToEvents: wf.LogToEvents,
		ReadOnly:    wf.ReadOnly,
		UpdatedAt:   wf.UpdatedAt,
		Namespace:   wf.Edges.Namespace.ID,
		Inode:       wf.Edges.Inode.ID,
		Refs:        refs,
		Revisions:   revisions,
		Routes:      routes,
	}
}

func entRef(ref *ent.Ref) *database.Ref {
	if ref == nil {
		return nil
	}

	x := &database.Ref{
		ID:        ref.ID,
		Name:      ref.Name,
		Immutable: ref.Immutable,
		CreatedAt: ref.CreatedAt,
	}

	if ref.Edges.Revision != nil {
		x.Revision = ref.Edges.Revision.ID
	}

	return x
}

func entRevision(rev *ent.Revision) *database.Revision {
	if rev == nil {
		return nil
	}

	return &database.Revision{
		ID:        rev.ID,
		CreatedAt: rev.CreatedAt,
		Hash:      rev.Hash,
		Source:    rev.Source,
		Metadata:  rev.Metadata,
		Workflow:  rev.Edges.Workflow.ID,
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
	}

	if inst.Edges.Namespace != nil {
		x.Namespace = inst.Edges.Namespace.ID
	}

	if inst.Edges.Workflow != nil {
		x.Workflow = inst.Edges.Workflow.ID
	}

	if inst.Edges.Revision != nil {
		x.Revision = inst.Edges.Revision.ID
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

func entMirror(v *ent.Mirror) *database.Mirror {
	if v == nil {
		return nil
	}

	return &database.Mirror{
		ID:         v.ID,
		URL:        v.URL,
		Ref:        v.Ref,
		Cron:       v.Cron,
		PublicKey:  v.PublicKey,
		PrivateKey: v.PrivateKey,
		Passphrase: v.Passphrase,
		Commit:     v.Commit,
		LastSync:   v.LastSync,
		UpdatedAt:  v.UpdatedAt,
		Inode:      v.Edges.Inode.ID,
	}
}

func entMirrorActivity(v *ent.MirrorActivity) *database.MirrorActivity {
	if v == nil {
		return nil
	}

	return &database.MirrorActivity{
		ID:         v.ID,
		Type:       v.Type,
		Status:     v.Status,
		CreatedAt:  v.CreatedAt,
		UpdatedAt:  v.UpdatedAt,
		EndAt:      v.EndAt,
		Controller: v.Controller,
		Deadline:   v.Deadline,
		Mirror:     v.Edges.Mirror.ID,
		Namespace:  v.Edges.Namespace.ID,
	}
}
