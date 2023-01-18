package mongo

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"

	"github.com/bldsoft/gost/cache"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type cacheChangeHandler[T any, U repository.IEntityIDPtr[T]] struct {
	cache          cache.ILocalCacheRepository
	cacheKeyPrefix string
}

func newCacheChangeHandler[T any, U repository.IEntityIDPtr[T]](cache cache.ILocalCacheRepository, cacheKeyPrefix string) *cacheChangeHandler[T, U] {
	return &cacheChangeHandler[T, U]{
		cache:          cache,
		cacheKeyPrefix: cacheKeyPrefix,
	}
}

func (h *cacheChangeHandler[T, U]) OnWarmUp(entities []U) {
	if err := h.CacheSet(entities...); err != nil {
		log.ErrorWithFields(log.Fields{"err": err}, "failed to warm up local cache")
	}
}

func (h *cacheChangeHandler[T, U]) OnChange(upd *UpdateEvent[T, U]) {
	switch upd.OpType {
	case Insert:
		fallthrough
	case Update:
		if err := h.CacheSet(upd.Entity); err != nil {
			log.Logger.ErrorWithFields(log.Fields{"err": err, "entity": upd.Entity}, "failed to update cache value")
		}
	case Delete:
		if err := h.CacheDelete(upd.Entity); err != nil {
			log.Logger.InfoWithFields(log.Fields{"err": err, "cache key": h.cacheKey(upd.Entity.StringID())}, "failed to delete cache value")
		}
	}
}

func (h cacheChangeHandler[T, U]) cacheMarshal(e U) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(e); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (h cacheChangeHandler[T, U]) cacheUnmarshal(data []byte) (U, error) {
	dec := gob.NewDecoder(bytes.NewBuffer(data))
	var e T
	if err := dec.Decode(&e); err != nil {
		return nil, err
	}
	return &e, nil
}

func (h cacheChangeHandler[T, U]) cacheKey(id string) string {
	return fmt.Sprintf("%s:%s", h.cacheKeyPrefix, id)
}

func (h cacheChangeHandler[T, U]) CacheSet(entities ...U) error {
	for _, e := range entities {
		if e == nil {
			continue
		}
		data, err := h.cacheMarshal(e)
		if err != nil {
			return err
		}
		if err := h.cache.Set(h.cacheKey(e.StringID()), data); err != nil {
			return err
		}
	}
	return nil
}

func (h cacheChangeHandler[T, U]) CacheDelete(e U) error {
	return h.cache.Delete(h.cacheKey(e.StringID()))
}

func (h cacheChangeHandler[T, U]) CacheGet(id string) (U, error) {
	data, err := h.cache.Get(h.cacheKey(id))
	if err != nil {
		return nil, err
	}
	return h.cacheUnmarshal(data)
}

type CachedRepositoryOptions struct {
	CacheKeyPrefix string
	WarmUp         bool
}

// CachedRepository is a wrapper for Repository, that keeps the entire collection in cache, updating it with Watcher.
// Only FindByID, FindByStringIDs, FindByIDs return cached results
type CachedRepository[T any, U repository.IEntityIDPtr[T]] struct {
	*WatchedRepository[T, U]
	cache *cacheChangeHandler[T, U]
}

func NewCachedRepository[T any, U repository.IEntityIDPtr[T]](db *Storage, collectionName string, cache cache.ILocalCacheRepository, opt ...CachedRepositoryOptions) *CachedRepository[T, U] {
	gob.Register(primitive.NilObjectID)
	var options CachedRepositoryOptions
	if len(opt) > 0 {
		options = opt[0]
	}
	changeHandler := newCacheChangeHandler[T, U](cache, options.CacheKeyPrefix)
	return &CachedRepository[T, U]{
		WatchedRepository: NewWatchedRepository[T, U](db, collectionName, changeHandler, options.WarmUp),
		cache:             changeHandler,
	}
}

func (r *CachedRepository[T, U]) cacheFindByID(ctx context.Context, id string, options ...*repository.QueryOptions) U {
	strID := repository.ToStringID[T, U](id)
	e, err := r.cache.CacheGet(strID)
	if err != nil {
		log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err, "collection": r.Repository.Name(), "id": strID}, "failed to get entity from cache")
	}
	return e
}

func (r *CachedRepository[T, U]) cacheFindByIDs(ctx context.Context, ids []string, options ...*repository.QueryOptions) []U {
	cachedRes := make([]U, 0, len(ids))
	for _, id := range ids {
		if e := r.cacheFindByID(ctx, id, options...); e != nil {
			cachedRes = append(cachedRes, e)
		} else {
			return nil
		}
	}
	return cachedRes
}

func (r *CachedRepository[T, U]) FindByID(ctx context.Context, id interface{}, options ...*repository.QueryOptions) (U, error) {
	if e := r.cacheFindByID(ctx, repository.ToStringID[T, U](id), options...); e != nil {
		log.FromContext(ctx).TraceWithFields(log.Fields{"collection": r.Repository.Name()}, "cache hit")
		return e, nil
	}

	res, err := r.Repository.FindByID(ctx, id, options...)
	if err != nil {
		return nil, err
	}

	if err := r.cache.CacheSet(res); err != nil {
		log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err}, "CachedRepository: failed to cache entity")
	}

	return res, nil
}

func (r *CachedRepository[T, U]) FindByStringIDs(ctx context.Context, ids []string, preserveOrder bool, options ...*repository.QueryOptions) ([]U, error) {
	if cachedRes := r.cacheFindByIDs(ctx, ids, options...); cachedRes != nil {
		log.FromContext(ctx).TraceWithFields(log.Fields{"collection": r.Repository.Name()}, "cache hit")
		return cachedRes, nil
	}

	res, err := r.Repository.FindByStringIDs(ctx, ids, preserveOrder, options...)
	if err != nil {
		return nil, err
	}

	if err := r.cache.CacheSet(res...); err != nil {
		log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err}, "CachedRepository: failed to cache entity")
	}

	return res, nil
}

func (r *CachedRepository[T, U]) FindByIDs(ctx context.Context, ids []interface{}, preserveOrder bool, options ...*repository.QueryOptions) ([]U, error) {
	if cachedRes := r.cacheFindByIDs(ctx, repository.ToStringIDs[T, U](ids), options...); cachedRes != nil {
		log.FromContext(ctx).TraceWithFields(log.Fields{"collection": r.Repository.Name()}, "cache hit")
		return cachedRes, nil
	}

	res, err := r.Repository.FindByIDs(ctx, ids, preserveOrder, options...)
	if err != nil {
		return nil, err
	}

	if err := r.cache.CacheSet(res...); err != nil {
		log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err}, "CachedRepository: failed to cache entity")
	}

	return res, nil
}

func (r *CachedRepository[T, U]) Delete(ctx context.Context, id interface{}, options ...*repository.QueryOptions) error {
	if err := r.cache.cache.Delete(r.cache.cacheKey(id.(string))); err != nil {
		log.Logger.InfoWithFields(log.Fields{"err": err, "cache key": id}, "failed to delete cache value")
	}

	return r.Repository.Delete(ctx, id, options...)
}
