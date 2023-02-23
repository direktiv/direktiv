package database

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/eko/gocache/lib/v4/store"
	"github.com/google/uuid"
)

func (db *CachedDatabase) lookupNamespaceByID(ctx context.Context, id uuid.UUID) *Namespace {

	key := fmt.Sprintf("nsid:%s", id)

	data, err := db.cache.Get(ctx, key)
	if err != nil {

		if !strings.Contains(err.Error(), "value not found in store") {
			db.sugar.Warnf("Namespace cache error: %v", err)
		}

		return nil

	}

	ns := new(Namespace)
	err = json.Unmarshal(data, ns)
	if err != nil {
		return nil
	}

	return ns

}

func (db *CachedDatabase) lookupNamespaceByName(ctx context.Context, name string) *Namespace {

	key := fmt.Sprintf("ns:%s", name)

	data, err := db.cache.Get(ctx, key)
	if err != nil {

		if !strings.Contains(err.Error(), "value not found in store") {
			db.sugar.Warnf("Namespace cache error: %v", err)
		}

		return nil

	}

	ns := new(Namespace)
	err = json.Unmarshal(data, ns)
	if err != nil {
		return nil
	}

	return ns

}

func (db *CachedDatabase) storeNamespaceInCache(ctx context.Context, ns *Namespace) {

	data, err := json.Marshal(ns)
	if err != nil {
		db.sugar.Warnf("Namespace cache marshal error: %v", err)
		return
	}

	key := fmt.Sprintf("nsid:%s", ns.ID)
	err = db.cache.Set(ctx, key, data, store.WithTags([]string{ns.ID.String()}))
	if err != nil {
		db.sugar.Warnf("Namespace cache store error: %v", err)
		return
	}

	key = fmt.Sprintf("ns:%s", ns.Name)
	err = db.cache.Set(ctx, key, data, store.WithTags([]string{ns.ID.String()}))
	if err != nil {
		db.sugar.Warnf("Namespace cache store error: %v", err)
		return
	}

}

func (db *CachedDatabase) invalidateCachedNamespace(ctx context.Context, id uuid.UUID, recursive bool) {

	if recursive {
		err := db.cache.Invalidate(ctx, store.WithInvalidateTags([]string{id.String()}))
		if err != nil {
			db.sugar.Error(err)
			return
		}
	} else {
		key := fmt.Sprintf("nsid:%s", id.String())
		err := db.cache.Delete(ctx, key)
		if err != nil {
			db.sugar.Error(err)
			return
		}
	}

}

func (db *CachedDatabase) lookupInodeByID(ctx context.Context, id uuid.UUID) *Inode {

	key := fmt.Sprintf("inoid:%s", id)

	data, err := db.cache.Get(ctx, key)
	if err != nil {

		if !strings.Contains(err.Error(), "value not found in store") {
			db.sugar.Warnf("Namespace cache error: %v", err)
		}

		return nil

	}

	ns := new(Inode)
	err = json.Unmarshal(data, ns)
	if err != nil {
		return nil
	}

	return ns

}

func (db *CachedDatabase) storeInodeInCache(ctx context.Context, ino *Inode) {

	data, err := json.Marshal(ino)
	if err != nil {
		db.sugar.Warnf("Inode cache marshal error: %v", err)
		return
	}

	key := fmt.Sprintf("inoid:%s", ino.ID)
	err = db.cache.Set(ctx, key, data, store.WithTags([]string{ino.Namespace.String()}))
	if err != nil {
		db.sugar.Warnf("Inode cache store error: %v", err)
		return
	}

}

func (db *CachedDatabase) invalidateCachedInode(ctx context.Context, id uuid.UUID, recursive bool) {

	if recursive {
		panic("TODO")
	}

	key := fmt.Sprintf("inoid:%s", id)

	err := db.cache.Delete(ctx, key)
	if err != nil {
		db.sugar.Error(err)
		return
	}

}

func (db *CachedDatabase) lookupWorkflowByID(ctx context.Context, id uuid.UUID) *Workflow {

	key := fmt.Sprintf("wfid:%s", id)

	data, err := db.cache.Get(ctx, key)
	if err != nil {

		if !strings.Contains(err.Error(), "value not found in store") {
			db.sugar.Warnf("Workflow cache error: %v", err)
		}

		return nil

	}

	wf := new(Workflow)
	err = json.Unmarshal(data, wf)
	if err != nil {
		return nil
	}

	return wf

}

func (db *CachedDatabase) storeWorkflowInCache(ctx context.Context, wf *Workflow) {

	data, err := json.Marshal(wf)
	if err != nil {
		db.sugar.Warnf("Workflow cache marshal error: %v", err)
		return
	}

	key := fmt.Sprintf("wfid:%s", wf.ID)
	err = db.cache.Set(ctx, key, data, store.WithTags([]string{wf.Namespace.String()}))
	if err != nil {
		db.sugar.Warnf("Workflow cache store error: %v", err)
		return
	}

}

func (db *CachedDatabase) invalidateCachedWorkflow(ctx context.Context, id uuid.UUID, recursive bool) {

	if recursive {
		panic("TODO")
	}

	key := fmt.Sprintf("wfid:%s", id)

	err := db.cache.Delete(ctx, key)
	if err != nil {
		db.sugar.Error(err)
		return
	}

}

func (db *CachedDatabase) lookupInstanceByID(ctx context.Context, id uuid.UUID) *Instance {

	key := fmt.Sprintf("instid:%s", id)

	data, err := db.cache.Get(ctx, key)
	if err != nil {

		if !strings.Contains(err.Error(), "value not found in store") {
			db.sugar.Warnf("Instance cache error: %v", err)
		}

		return nil

	}

	inst := new(Instance)
	err = json.Unmarshal(data, inst)
	if err != nil {
		return nil
	}

	return inst

}

func (db *CachedDatabase) storeInstanceInCache(ctx context.Context, inst *Instance) {

	data, err := json.Marshal(inst)
	if err != nil {
		db.sugar.Warnf("Instance cache marshal error: %v", err)
		return
	}

	key := fmt.Sprintf("instid:%s", inst.ID)
	err = db.cache.Set(ctx, key, data, store.WithTags([]string{inst.Namespace.String()}))
	if err != nil {
		db.sugar.Warnf("Instance cache store error: %v", err)
		return
	}

}

func (db *CachedDatabase) lookupRevisionByID(ctx context.Context, id uuid.UUID) *Revision {

	key := fmt.Sprintf("revid:%s", id)

	data, err := db.cache.Get(ctx, key)
	if err != nil {

		if !strings.Contains(err.Error(), "value not found in store") {
			db.sugar.Warnf("Revision cache error: %v", err)
		}

		return nil

	}

	rev := new(Revision)
	err = json.Unmarshal(data, rev)
	if err != nil {
		return nil
	}

	return rev

}

func (db *CachedDatabase) storeRevisionInCache(ctx context.Context, rev *Revision) {

	data, err := json.Marshal(rev)
	if err != nil {
		db.sugar.Warnf("Revision cache marshal error: %v", err)
		return
	}

	key := fmt.Sprintf("revid:%s", rev.ID)
	err = db.cache.Set(ctx, key, data) // , store.WithTags([]string{rev.Namespace.String()}))
	if err != nil {
		db.sugar.Warnf("Revision cache store error: %v", err)
		return
	}

}
