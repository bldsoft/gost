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
func (r *MemcacheRepository) Get(key string) ([]byte, error) {
	key = r.cache.PrepareKey(key)
	item, err := r.cache.Get(key)
	if err != nil || item == nil {
		return nil, r.mapError(err)
	}
	return item.Value, err
}

func (r *MemcacheRepository) GetWithFlags(key string) (data []byte, flags uint32, err error) {
	key = r.cache.PrepareKey(key)
	item, err := r.cache.Get(key)
	if err != nil || item == nil {
		return nil, 0, r.mapError(err)
	}
	return item.Value, item.Flags, err
}

func (r *MemcacheRepository) GetMulti(keys []string) (map[string][]byte, error) {
	keys = r.cache.PrepareKeys(keys)
	items, err := r.cache.GetMulti(keys)
	m := make(map[string][]byte, len(items))
	for key, item := range items {
		m[key] = item.Value
	}
	return m, err
}

// SetLiveTimeMin ...
func (r *MemcacheRepository) SetLiveTimeMin(liveTime time.Duration) {
	r.liveTime = liveTime
}

// Set writes the given item, unconditionally.
func (r *MemcacheRepository) Set(key string, value []byte) error {
	return r.SetFor(key, value, r.liveTime)
}

func (r *MemcacheRepository) truncExpiration(d time.Duration) int32 {
	const maxDuration = 30 * 24 * time.Hour
	if d > maxDuration {
		return int32(maxDuration.Seconds())
	}
	return int32(d.Seconds())
}

// SetFor writes the given item, unconditionally.
func (r *MemcacheRepository) SetFor(key string, value []byte, expiration time.Duration) error {
	key = r.cache.PrepareKey(key)
	return r.cache.Set(&memcache.Item{Key: key, Value: value, Expiration: r.truncExpiration(expiration)})
}

func (r *MemcacheRepository) SetWithFlags(key string, value []byte, flags uint32) error {
	key = r.cache.PrepareKey(key)
	return r.cache.Set(&memcache.Item{Key: key, Value: value, Flags: flags, Expiration: int32(r.liveTime.Seconds())})
}

func (r *MemcacheRepository) SetForWithFlags(key string, value []byte, flags uint32, expiration time.Duration) error {
	key = r.cache.PrepareKey(key)
	return r.cache.Set(&memcache.Item{Key: key, Value: value, Flags: flags, Expiration: r.truncExpiration(expiration)})
}

// Exist checks if the key exists
func (r *MemcacheRepository) Exist(key string) bool {
	key = r.cache.PrepareKey(key)
	return r.cache.Touch(key, int32(r.liveTime.Seconds())) == nil
}

// Add writes the given item, if no value already exists for its
// key. ErrNotStored is returned if that condition is not met.
func (r *MemcacheRepository) Add(key string, value []byte) error {
	return r.AddFor(key, value, r.liveTime)
}

// AddFor writes the given item, if no value already exists for its
// key. ErrExists is returned if that condition is not met.
func (r *MemcacheRepository) AddFor(key string, value []byte, expiration time.Duration) error {
	key = r.cache.PrepareKey(key)
	err := r.cache.Add(&memcache.Item{Key: key, Value: value, Expiration: r.truncExpiration(expiration)})
	if errors.Is(err, memcache.ErrNotStored) {
		return cache.ErrExists
	}
	return err
}

func (r *MemcacheRepository) AddWithFlags(key string, value []byte, flags uint32) error {
	key = r.cache.PrepareKey(key)
	err := r.cache.Add(&memcache.Item{Key: key, Value: value, Flags: flags, Expiration: int32(r.liveTime.Seconds())})
	if errors.Is(err, memcache.ErrNotStored) {
		return cache.ErrExists
	}
	return err
}

func (r *MemcacheRepository) AddForWithFlags(key string, value []byte, flags uint32, expiration time.Duration) error {
	key = r.cache.PrepareKey(key)
	err := r.cache.Add(&memcache.Item{Key: key, Value: value, Flags: flags, Expiration: r.truncExpiration(expiration)})
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
