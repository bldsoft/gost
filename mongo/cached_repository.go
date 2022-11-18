package mongo

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"

	"github.com/bldsoft/gost/cache"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/repository"
	"go.mongodb.org/mongo-driver/bson"
)

const updateChanBufferSize = 12500

type update[T any, U repository.IEntityIDPtr[T]] struct {
	e      U
	opType OperationType
}

// CachedRepository is a wrapper for Repository, that keeps the entire collection in cache, updating it with Watcher.
// Only FindByID, FindByStringIDs, FindByIDs return cached results
type CachedRepository[T any, U repository.IEntityIDPtr[T]] struct {
	*Repository[T, U]
	cache       cache.ILocalCacheRepository
	watcher     *Watcher
	cachePrefix string
}

func NewCachedRepository[T any, U repository.IEntityIDPtr[T]](db *Storage, collectionName string, cache cache.ILocalCacheRepository, cachePrefix string) *CachedRepository[T, U] {
	rep := &CachedRepository[T, U]{
		Repository:  NewRepository[T, U](db, collectionName),
		cache:       cache,
		cachePrefix: cachePrefix,
	}
	rep.init()
	return rep
}

func (r *CachedRepository[T, U]) init() {
	updateC := make(chan *update[T, U], updateChanBufferSize)

	r.watcher = NewWatcher(r.Repository.Collection())
	r.watcher.SetHandler(func(fullDocument bson.Raw, opType OperationType) {
		var e T
		if err := bson.Unmarshal(fullDocument, &e); err != nil {
			log.Logger.ErrorWithFields(log.Fields{"err": err}, "failed to update %T in cache")
		}
		updateC <- &update[T, U]{
			e:      &e,
			opType: opType,
		}
	})
	r.watcher.Start()

	if err := r.warmUpCache(context.Background()); err != nil {
		log.Logger.ErrorWithFields(log.Fields{"err": err, "collection": r.Repository.collectionName}, "Failed to warm up cache")
	} else {
		log.Logger.DebugWithFields(log.Fields{"collection": r.Repository.collectionName}, "Cache is warmed up")
	}

	go func() {
		for upd := range updateC {
			switch upd.opType {
			case Insert:
				fallthrough
			case Update:
				if err := r.cacheSet(upd.e); err != nil {
					log.Logger.ErrorWithFields(log.Fields{"err": err, "entity": upd.e}, "failed to update cache value")
				}
			case Delete:
				if err := r.cacheDelete(upd.e); err != nil {
					log.Logger.ErrorWithFields(log.Fields{"err": err, "cache key": r.cacheKey(upd.e.StringID())}, "failed to delete cache value")
				}
			}
		}
	}()
}

func (r *CachedRepository[T, U]) cacheMarshal(e U) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(e); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (r *CachedRepository[T, U]) cacheUnmarshal(data []byte) (U, error) {
	dec := gob.NewDecoder(bytes.NewBuffer(data))
	var e T
	if err := dec.Decode(&e); err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *CachedRepository[T, U]) cacheKey(id string) string {
	return fmt.Sprintf("%s:%s", r.cachePrefix, id)
}

func (r *CachedRepository[T, U]) cacheSet(e U) error {
	data, err := r.cacheMarshal(e)
	if err != nil {
		return err
	}
	return r.cache.Set(r.cacheKey(e.StringID()), data)
}

func (r *CachedRepository[T, U]) cacheDelete(e U) error {
	return r.cache.Delete(r.cacheKey(e.StringID()))
}

func (r *CachedRepository[T, U]) cacheGet(id string) (U, error) {
	data, err := r.cache.Get(r.cacheKey(id))
	if err != nil {
		return nil, err
	}
	return r.cacheUnmarshal(data)
}

func (r *CachedRepository[T, U]) warmUpCache(ctx context.Context) error {
	entities, err := r.Repository.GetAll(context.Background())
	if err != nil {
		return err
	}
	for _, e := range entities {
		if err = r.cacheSet(e); err != nil {
			return err
		}
	}
	return nil
}

func (r *CachedRepository[T, U]) cacheFindByID(ctx context.Context, id string, options ...*repository.QueryOptions) U {
	strID := repository.ToStringID[T, U](id)
	e, err := r.cacheGet(strID)
	if err != nil {
		log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err, "collection": r.Repository.collectionName, "id": strID}, "failed to get entity from cache")
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
		return e, nil
	}
	return r.Repository.FindByID(ctx, id, options...)
}

func (r *CachedRepository[T, U]) FindByStringIDs(ctx context.Context, ids []string, preserveOrder bool, options ...*repository.QueryOptions) ([]U, error) {
	if cachedRes := r.cacheFindByIDs(ctx, ids, options...); cachedRes != nil {
		return cachedRes, nil
	}
	return r.Repository.FindByStringIDs(ctx, ids, preserveOrder, options...)
}

func (r *CachedRepository[T, U]) FindByIDs(ctx context.Context, ids []interface{}, preserveOrder bool, options ...*repository.QueryOptions) ([]U, error) {
	if cachedRes := r.cacheFindByIDs(ctx, repository.ToStringIDs[T, U](ids), options...); cachedRes != nil {
		return cachedRes, nil
	}
	return r.Repository.FindByIDs(ctx, ids, preserveOrder, options...)
}
