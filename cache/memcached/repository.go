package memcached

import (
	"errors"
	"time"

	"github.com/bldsoft/gost/cache"
	"github.com/bradfitz/gomemcache/memcache"
)

const (
	casRetryLimit = 5
	casSleepTime  = 10
)

type MemcacheRepository struct {
	cache    *Storage
	liveTime time.Duration
}

func NewMemcacheRepository(storage *Storage, liveTime time.Duration) *MemcacheRepository {
	rep := &MemcacheRepository{cache: storage}
	rep.SetLiveTimeMin(liveTime)
	return rep
}

// Get gets the item valut for the given key. ErrCacheMiss is returned for a
// memcache cache miss. The key must be at most 250 bytes in length.
func (r *MemcacheRepository) Get(key string) (*cache.Item, error) {
	key = r.cache.PrepareKey(key)
	item, err := r.cache.Get(key)
	if err != nil || item == nil {
		return nil, r.mapError(err)
	}
	return &cache.Item{
		Value: item.Value,
		// TTL:   time.Duration(item.Expiration) * time.Second,
		Flags: item.Flags,
	}, err
}

func (r *MemcacheRepository) Exist(key string) bool {
	key = r.cache.PrepareKey(key)
	return r.cache.Touch(key, int32(r.liveTime.Seconds())) == nil
}

// SetLiveTimeMin ...
func (r *MemcacheRepository) SetLiveTimeMin(liveTime time.Duration) {
	r.liveTime = liveTime
}

// Set writes the given item, unconditionally.
func (r *MemcacheRepository) Set(key string, val []byte, item ...cache.ItemF) error {
	return r.cache.Set(r.item(key, val, item...))
}

func (r *MemcacheRepository) Add(key string, val []byte, item ...cache.ItemF) error {
	err := r.cache.Add(r.item(key, val, item...))
	if errors.Is(err, memcache.ErrNotStored) {
		return cache.ErrExists
	}

	return err
}

// Delete deletes the item with the provided key.
func (r *MemcacheRepository) Delete(key string) error {
	key = r.cache.PrepareKey(key)
	return r.mapError(r.cache.Delete(key))
}

// Reset ...
func (r *MemcacheRepository) Reset() {
	r.cache.FlushAll()
}

func (r *MemcacheRepository) CompareAndSwap(key string, handler func(value *cache.Item) (*cache.Item, error), sleepDur ...time.Duration) error {
	var err error
	key = r.cache.PrepareKey(key)

	for i := 0; i < casRetryLimit; i++ {
		item, err := r.cache.Get(key)

		if err != nil || item == nil {
			return err
		}

		data, err := handler(&cache.Item{
			Value: item.Value,
			Flags: item.Flags,
		})

		if err != nil || data == nil {
			return err
		}

		item.Value = data.Value
		if data.Flags != 0 {
			item.Flags = data.Flags
		}
		err = r.cache.CompareAndSwap(item)

		switch err {
		case memcache.ErrCASConflict:
			time.Sleep(casSleepTime * time.Millisecond)
		default:
			return err
		}
	}

	return err
}

// retry on failed get
// return bool for set
func (r *MemcacheRepository) AddOrGet(key string, val []byte, opts ...cache.ItemF) (*cache.Item, bool, error) {
	i, err := r.Get(key)
	if err == nil {
		return &cache.Item{
			Value: i.Value,
			Flags: i.Flags,
		}, false, nil
	}

	if err := r.Add(key, val, opts...); errors.Is(err, cache.ErrExists) {
		i, err := r.Get(key)
		return i, false, err
	}

	return nil, true, err
}

func (r *MemcacheRepository) mapError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, memcache.ErrCacheMiss):
		return cache.ErrCacheMiss
	default:
		return err
	}
}

func (r *MemcacheRepository) item(key string, val []byte, itemFs ...cache.ItemF) *memcache.Item {
	it := memcache.Item{
		Key:        r.cache.PrepareKey(key),
		Value:      val,
		Expiration: truncExpiration(r.liveTime),
	}

	if len(itemFs) == 0 {
		return &it
	}

	cIt := &cache.Item{}
	for _, f := range itemFs {
		f(cIt)
	}
	if cIt.Flags != 0 {
		it.Flags = cIt.Flags
	}
	if cIt.TTL != 0 {
		it.Expiration = truncExpiration(cIt.TTL)
	}
	return &it
}

// func (r *MemcacheRepository) item(key string, item *cache.Item) *memcache.Item {
// 	it := memcache.Item{
// 		Key:        r.cache.PrepareKey(key),
// 		Expiration: truncExpiration(r.liveTime),
// 	}

// 	if item != nil {
// 		it.Value = item.Value
// 		if item.TTL != 0 {
// 			it.Expiration = truncExpiration(item.TTL)
// 		}
// 		if item.Flags != 0 {
// 			it.Flags = item.Flags
// 		}
// 	}

// 	return &it
// }

func truncExpiration(d time.Duration) int32 {
	const maxDuration = 30 * 24 * time.Hour
	if d > maxDuration {
		return int32(maxDuration.Seconds())
	}
	return int32(d.Seconds())
}

var _ (cache.IDistrCacheRepository) = NewMemcacheRepository(nil, time.Second)
