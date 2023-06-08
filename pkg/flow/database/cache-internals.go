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
	if !db.cachingEnabled {
		return nil
	}

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
	if !db.cachingEnabled {
		return nil
	}

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
	if !db.cachingEnabled {
		return
	}

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
	if !db.cachingEnabled {
		return
	}

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
