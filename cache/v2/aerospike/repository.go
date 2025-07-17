package aerospike

import (
	"errors"
	"fmt"
	"time"

	"github.com/bldsoft/gost/cache/v2"
	"github.com/bldsoft/gost/log"

	aero "github.com/aerospike/aerospike-client-go/v8"
	aeroTypes "github.com/aerospike/aerospike-client-go/v8/types"
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
	setName  string
}

func NewRepository(cache *Storage, liveTime time.Duration, setName string) *Repository {
	rep := &Repository{cache: cache, setName: setName}
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
	item, err := r.cache.Get(r.cache.getReadPolicy(), asKey)
	if err != nil || item == nil {
		if errors.Is(err, aero.ErrKeyNotFound) {
			return nil, cache.ErrCacheMiss
		}
		return nil, err
	}
	res := &cache.Item{}
	res.Value = item.Bins[valueBinKey].([]byte)
	res.TTL = time.Duration(item.Expiration) * time.Second
	if flags, ok := item.Bins[flagsBinKey]; ok {
		res.Flags = uint32(flags.(int))
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
	err = r.cache.PutBins(p, asKey, b...)
	if err == nil {
		return nil
	}
	var asErr *aero.AerospikeError
	if errors.As(err, &asErr) && asErr.ResultCode == aeroTypes.KEY_EXISTS_ERROR {
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
	return r.cache.PutBins(p, asKey, b...)
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
		item, aErr := r.cache.Get(r.cache.getReadPolicy(), asKey)
		if aErr != nil {
			return aErr
		}
		if item == nil {
			return cache.ErrCacheMiss
		}

		data := &cache.Item{}
		if val, ok := item.Bins[valueBinKey]; ok {
			data.Value = val.([]byte)
		}
		if flags, ok := item.Bins[flagsBinKey]; ok {
			data.Flags = uint32(flags.(int))
		}

		data, err := handler(data)
		if err != nil || data == nil {
			return err
		}
		item.Bins[valueBinKey] = data.Value
		if data.Flags != 0 {
			item.Bins[flagsBinKey] = data.Flags
		}

		wp := r.cache.getWritePolicy(item.Generation, 0)
		wp.GenerationPolicy = aero.EXPECT_GEN_EQUAL

		err = r.cache.Put(wp, asKey, item.Bins)
		var asErr *aero.AerospikeError
		if errors.As(err, &asErr) && asErr.ResultCode == aeroTypes.GENERATION_ERROR {
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

func (r *Repository) Reset() {
	now := time.Now()
	if err := r.cache.Truncate(nil, r.cache.namespace, r.setName, &now); err != nil {
		log.WarnWithFields(log.Fields{"err": err}, "failed to reset aerospike cache")
	}
}

func (r *Repository) key(key string) (*aero.Key, error) {
	return aero.NewKey(r.cache.namespace, r.setName, key)
}

func (r *Repository) item(replace bool, val []byte, itemFs ...cache.ItemF) (*aero.WritePolicy, []*aero.Bin) {
	bins := []*aero.Bin{aero.NewBin(valueBinKey, val)}
	policy := r.cache.getWritePolicy(0, truncExpiration(r.liveTime))
	policy.RecordExistsAction = aero.CREATE_ONLY
	if replace {
		policy.RecordExistsAction = aero.UPDATE
	}

	if len(itemFs) == 0 {
		return policy, bins
	}
	it := cache.CollectItem(itemFs...)
	if it.Flags != 0 {
		bins = append(bins, aero.NewBin(flagsBinKey, it.Flags))
	}
	if it.TTL != 0 {
		policy.Expiration = truncExpiration(it.TTL)
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
