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
func (r *MemcacheRepository) Get(key string) ([]byte, uint32, error) {
	key = r.cache.PrepareKey(key)
	item, err := r.cache.Get(key)
	if err != nil || item == nil {
		return nil, 0, r.mapError(err)
	}
	return item.Value, item.Flags, err
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
func (r *MemcacheRepository) Set(key string, opts *cache.Options) error {
	return r.cache.Set(r.item(key, opts))
}

func (r *MemcacheRepository) Add(key string, opts *cache.Options) error {
	err := r.cache.Add(r.item(key, opts))
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

func (r *MemcacheRepository) CompareAndSwap(key string, handler func(value []byte) ([]byte, error)) error {
	var err error
	key = r.cache.PrepareKey(key)

	for i := 0; i < casRetryLimit; i++ {
		item, err := r.cache.Get(key)

		if err != nil || item == nil {
			return err
		}

		data, err := handler(item.Value)

		if err != nil || data == nil {
			return err
		}

		item.Value = data
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

func (r *MemcacheRepository) item(key string, opts *cache.Options) *memcache.Item {
	it := memcache.Item{
		Key:        r.cache.PrepareKey(key),
		Expiration: truncExpiration(r.liveTime),
	}

	if opts != nil {
		it.Value = opts.Value
		if opts.TTL != nil {
			it.Expiration = truncExpiration(*opts.TTL)
		}
		if opts.Flags != nil {
			it.Flags = *opts.Flags
		}
	}

	return &it
}

func truncExpiration(d time.Duration) int32 {
	const maxDuration = 30 * 24 * time.Hour
	if d > maxDuration {
		return int32(maxDuration.Seconds())
	}
	return int32(d.Seconds())
}

var _ (cache.DistrCacheRepository) = NewMemcacheRepository(nil, time.Second)
