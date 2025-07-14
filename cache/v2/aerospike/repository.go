package aerospike

import (
	"errors"
	"fmt"
	"time"

	"github.com/bldsoft/gost/cache/v2"

	aero "github.com/aerospike/aerospike-client-go/v8"
)

const (
	valueBinKey = "value"
	flagsBinKey = "flags"

	casRetryLimit = 5
	casSleepTime  = 10
)

type Repository struct {
	cache    *Storage
	liveTime time.Duration
}

func NewRepository(cache *Storage, liveTime time.Duration) *Repository {
	rep := &Repository{cache: cache}
	rep.SetLiveTimeMin(liveTime)
	return rep
}

func (r *Repository) SetLiveTimeMin(liveTime time.Duration) {
	r.liveTime = liveTime
}

func (r *Repository) Get(key string) (*cache.Item, error) {
	asKey, err := r.key(key)
	if err != nil {
		return nil, err
	}
	item, err := r.cache.Get(nil, asKey)
	if err != nil || item == nil {
		if errors.Is(err, aero.ErrKeyNotFound) {
			return nil, cache.ErrCacheMiss
		}
		return nil, err
	}
	res := &cache.Item{}
	res.Value = item.Bins[valueBinKey].([]byte)
	res.TTL = time.Duration(item.Expiration) * time.Second
	if flags, ok := item.Bins["flags"]; ok {
		res.Flags = flags.(uint32)
	}
	return res, nil
}

func (r *Repository) Exist(key string) bool {
	asKey, err := r.key(key)
	if err != nil {
		return false
	}
	exists, _ := r.cache.Exists(nil, asKey)
	return exists
}

func (r *Repository) Delete(key string) error {
	asKey, err := r.key(key)
	if err != nil {
		return err
	}
	_, err = r.cache.Delete(nil, asKey)
	return err
}

func (r *Repository) Add(key string, val []byte, item ...cache.ItemF) error {
	asKey, err := r.key(key)
	if err != nil {
		return err
	}
	p, b := r.item(false, val, item...)
	err = r.cache.Put(p, asKey, b)
	if err == nil {
		return nil
	}
	var asErr *aero.AerospikeError
	if errors.As(err, &asErr) && asErr.ResultCode == 2 {
		return cache.ErrExists
	}
	return err
}

func (r *Repository) Set(key string, val []byte, item ...cache.ItemF) error {
	asKey, err := r.key(key)
	if err != nil {
		return err
	}
	p, b := r.item(true, val, item...)
	return r.cache.Put(p, asKey, b)
}

func (r *Repository) CompareAndSwap(
	key string,
	handler func(value *cache.Item) (*cache.Item, error),
	sleepDur ...time.Duration,
) error {
	asKey, err := r.key(key)
	if err != nil {
		return err
	}

	for i := 0; i < casRetryLimit; i++ {
		item, aErr := r.cache.Get(nil, asKey)
		if aErr != nil {
			return aErr
		}
		if item == nil {
			return cache.ErrCacheMiss
		}

		newItem, err := handler(&cache.Item{
			Value: item.Bins[valueBinKey].([]byte),
			TTL:   time.Duration(item.Expiration) * time.Second,
		})
		if flags, ok := item.Bins["flags"]; ok {
			newItem.Flags = flags.(uint32)
		}
		if err != nil || newItem == nil {
			return err
		}

		wp := aero.NewWritePolicy(0, 0)
		wp.GenerationPolicy = aero.EXPECT_GEN_EQUAL
		wp.Generation = item.Generation

		bins := aero.BinMap{
			"value": newItem.Value,
			"flags": newItem.Flags,
		}

		err = r.cache.Put(wp, asKey, bins)

		var asErr *aero.AerospikeError
		if errors.As(err, &asErr) && asErr.ResultCode == 3 {
			time.Sleep(casSleepTime * time.Millisecond)
			continue
		}
		return err
	}

	return fmt.Errorf("aero: CAS retry limit exceeded")
}

func (r *Repository) AddOrGet(key string, val []byte, opts ...cache.ItemF) (*cache.Item, bool, error) {
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

func (r *Repository) Reset() {}

func (r *Repository) key(key string) (*aero.Key, error) {
	return aero.NewKey(r.cache.namespace, r.cache.keyPrefix, key)
}

func (r *Repository) item(replace bool, val []byte, itemFs ...cache.ItemF) (*aero.WritePolicy, aero.BinMap) {
	bins := aero.BinMap{
		valueBinKey: val,
	}
	policy := aero.NewWritePolicy(0, truncExpiration(r.liveTime))
	policy.RecordExistsAction = aero.CREATE_ONLY
	if replace {
		policy.RecordExistsAction = aero.UPDATE
	}
	if len(itemFs) == 0 {
		return policy, bins
	}
	it := cache.CollectItem(itemFs...)
	if it.Flags != 0 {
		bins[flagsBinKey] = it.Flags
	}
	if it.TTL != 0 {
		policy.Expiration = truncExpiration(it.TTL)
	}
	if len(it.Extras) > 0 {
		for k, v := range it.Extras {
			if k == flagsBinKey || k == valueBinKey {
				continue
			}
			bins[k] = v
		}
	}

	return policy, bins
}

func truncExpiration(d time.Duration) uint32 {
	const maxDuration = 30 * 24 * time.Hour
	if d > maxDuration {
		return uint32(maxDuration.Seconds())
	}
	return uint32(d.Seconds())
}

var _ cache.IDistrCacheRepository = (*Repository)(nil)
