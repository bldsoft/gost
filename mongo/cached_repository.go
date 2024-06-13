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

type cacheWatcher[T any, U repository.IEntityIDPtr[T]] struct {
	cache cache.ILocalCacheRepository
	opt   CachedRepositoryOptions
}

func newCacheWatcher[T any, U repository.IEntityIDPtr[T]](cache cache.ILocalCacheRepository, opt CachedRepositoryOptions) *cacheWatcher[T, U] {
	return &cacheWatcher[T, U]{
		cache: cache,
		opt:   opt,
	}
}

func (h *cacheWatcher[T, U]) WarmUp(ctx context.Context, rep repository.Repository[T, U]) error {
	if !h.opt.WarmUp {
		return nil
	}
	entities, err := rep.GetAll(ctx)
	if err != nil {
		return err
	}
	err = h.CacheSet(entities...)
	if err != nil {
		return fmt.Errorf("failed to cache entities: %w", err)
	}
	return nil
}

func (h *cacheWatcher[T, U]) OnEvent(upd repository.Event[T, U]) {
	switch upd.Type {
	case repository.EventTypeCreate:
		fallthrough
	case repository.EventTypeUpdate:
		if err := h.CacheSet(upd.Entity); err != nil {
			log.Logger.ErrorWithFields(log.Fields{"err": err, "entity": upd.Entity}, "failed to update cache value")
		}
	case repository.EventTypeDelete:
		if err := h.CacheDelete(upd.Entity.StringID()); err != nil {
			log.Logger.DebugWithFields(log.Fields{"err": err, "cache key": h.cacheKey(upd.Entity.StringID())}, "failed to delete cache value")
		}
	}
}

func (h cacheWatcher[T, U]) cacheMarshal(e U) ([]byte, error) {
	if marshal := h.opt.Marshal; marshal != nil {
		return marshal(e)
	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(e); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (h cacheWatcher[T, U]) cacheUnmarshal(data []byte) (U, error) {
	var e T
	var err error
	if unmarshal := h.opt.Unmarshal; unmarshal != nil {
		err = unmarshal(data, &e)
	} else {
		dec := gob.NewDecoder(bytes.NewBuffer(data))
		err = dec.Decode(&e)
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (h cacheWatcher[T, U]) cacheKey(id string) string {
	return fmt.Sprintf("%s:%s", h.opt.CacheKeyPrefix, id)
}

func (h cacheWatcher[T, U]) CacheSet(entities ...U) error {
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

func (h cacheWatcher[T, U]) CacheDelete(id string) error {
	return h.cache.Delete(h.cacheKey(id))
}

func (h cacheWatcher[T, U]) CacheGet(id string) (U, error) {
	data, err := h.cache.Get(h.cacheKey(id))
	if err != nil {
		return nil, err
	}
	return h.cacheUnmarshal(data)
}

type CachedRepositoryOptions struct {
	CacheKeyPrefix string
	WarmUp         bool
	Marshal        func(any) ([]byte, error)      // gob if nil
	Unmarshal      func(data []byte, v any) error // gob if nil
}

// CachedRepository is a wrapper for Repository, that keeps the entire collection in cache, updating it with Watcher.
// Only FindByID, FindByStringIDs, FindByIDs return cached results
type CachedRepository[T any, U repository.IEntityIDPtr[T]] struct {
	*WatchedRepository[T, U]
	cache *cacheWatcher[T, U]
}

func NewCachedRepository[T any, U repository.IEntityIDPtr[T]](db *Storage, collectionName string, cache cache.ILocalCacheRepository, opt ...CachedRepositoryOptions) *CachedRepository[T, U] {
	gob.Register(primitive.NilObjectID)
	var options CachedRepositoryOptions
	if len(opt) > 0 {
		options = opt[0]
	}
	cacheWatcher := newCacheWatcher[T, U](cache, options)
	return &CachedRepository[T, U]{
		WatchedRepository: NewWatchedRepository[T, U](db, collectionName, cacheWatcher),
		cache:             cacheWatcher,
	}
}

func (r *CachedRepository[T, U]) cacheFindByID(ctx context.Context, id string, options ...*repository.QueryOptions) U {
	strID := repository.ToStringID[T, U](id)
	e, err := r.cache.CacheGet(strID)
	if err != nil {
		log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err, "collection": r.Repository.Name(), "id": strID}, "failed to get entity from cache")
		return nil
	}
	if len(options) == 0 || options[0].Archived {
		return e
	}

	if withArchived, ok := any(e).(IEntityArchived); !ok {
		return e
	} else if options[0].Archived != withArchived.IsArchived() {
		return nil
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
	strID := repository.ToStringID[T, U](id)
	if err := r.cache.CacheDelete(strID); err != nil {
		log.Logger.DebugWithFields(log.Fields{"err": err, "cache key": id}, "failed to delete cache value")
	}

	return r.Repository.Delete(ctx, id, options...)
}
