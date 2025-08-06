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
	valueBinKey        = "value"
	flagsBinKey        = "flags"
	continuationBinKey = "continuation"
	lenBinKey          = "totalValueSize"

	casRetryLimit = 5
	casSleepTime  = 10

	defaultItemSizeLimit = int(1024 * 1024 * 6.5) // 6.5MB
)

type Repository struct {
	cache         *Storage
	liveTime      time.Duration
	setName       string
	itemSizeLimit int
}

func NewRepository(cache *Storage, liveTime time.Duration, setName string) *Repository {
	rep := &Repository{
		cache:         cache,
		setName:       setName,
		itemSizeLimit: defaultItemSizeLimit,
	}
	rep.SetLiveTimeMin(liveTime)
	return rep
}

func (r *Repository) SetLiveTimeMin(liveTime time.Duration) {
	r.liveTime = liveTime
}

func (r *Repository) SetItemSizeLimit(limit int) {
	if limit > 0 {
		r.itemSizeLimit = limit
	}
}

func (r *Repository) Get(key string) (*cache.Item, error) {
	res, _, err := r.get(key)
	return res, err
}

func (r *Repository) get(key string) (*cache.Item, uint32, error) {
	asKey, err := r.key(key)
	if err != nil {
		return nil, 0, err
	}
	item, err := r.cache.Get(r.cache.getReadPolicy(), asKey)
	if err != nil || item == nil {
		if errors.Is(err, aero.ErrKeyNotFound) {
			return nil, 0, cache.ErrCacheMiss
		}
		return nil, 0, err
	}

	res := &cache.Item{}
	res.TTL = time.Duration(item.Expiration) * time.Second
	if flags, ok := item.Bins[flagsBinKey]; ok {
		res.Flags = uint32(flags.(int))
	}

	mainValue := item.Bins[valueBinKey].([]byte)
	totalSize := len(mainValue)

	var continuationKeys []*aero.Key
	if item.Bins[continuationBinKey] != nil {
		continuationKeys = make([]*aero.Key, len(item.Bins[continuationBinKey].([]interface{})))
		for i, k := range item.Bins[continuationBinKey].([]interface{}) {
			asKey, err := r.key(k.(string))
			if err != nil {
				return nil, 0, err
			}
			continuationKeys[i] = asKey
		}

		totalSize += len(continuationKeys) * r.itemSizeLimit
		log.TraceWithFields(log.Fields{
			"key":                key,
			"mainSize":           len(mainValue),
			"continuations":      len(continuationKeys),
			"estimatedTotalSize": totalSize,
		}, "aerospike: reconstructing from continuations")
	} else {
		res.Value = mainValue
		return res, item.Generation, nil
	}

	if l, ok := item.Bins[lenBinKey]; ok {
		totalSize = l.(int)
	}
	res.Value = make([]byte, 0, totalSize)
	res.Value = append(res.Value, mainValue...)

	if len(continuationKeys) > 0 {
		continuations, err := r.cache.BatchGet(nil, continuationKeys, valueBinKey)
		if err != nil {
			return nil, 0, err
		}
		for _, c := range continuations {
			res.Value = append(res.Value, c.Bins[valueBinKey].([]byte)...)
		}
	}

	return res, item.Generation, nil
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

	item, err := r.cache.Get(r.cache.getReadPolicy(), asKey)
	if err != nil && !errors.Is(err, aero.ErrKeyNotFound) {
		return err
	}

	_, err = r.cache.Delete(nil, asKey)
	if err != nil {
		return err
	}

	if item != nil && item.Bins[continuationBinKey] != nil {
		continuationKeys := item.Bins[continuationBinKey].([]interface{})
		if len(continuationKeys) > 0 {
			keys := make([]*aero.Key, 0, len(continuationKeys))
			for _, k := range continuationKeys {
				asKey, err := r.key(k.(string))
				if err != nil {
					log.WarnWithFields(log.Fields{"key": k, "err": err}, "failed to create key for continuation deletion")
					continue
				}
				keys = append(keys, asKey)
			}

			if len(keys) > 0 {
				batchDeletes := make([]aero.BatchRecordIfc, 0, len(keys))
				for _, key := range keys {
					batchDeletes = append(batchDeletes, aero.NewBatchDelete(nil, key))
				}
				if err := r.cache.BatchOperate(nil, batchDeletes); err != nil {
					log.WarnWithFields(log.Fields{"err": err}, "failed to delete continuation records")
				}
			}
		}
	}

	return nil
}

func (r *Repository) Add(key string, val []byte, item ...cache.ItemF) error {
	err := r.put(false, key, val, nil, item...)
	var asErr *aero.AerospikeError
	if errors.As(err, &asErr) && asErr.ResultCode == aeroTypes.KEY_EXISTS_ERROR {
		return cache.ErrExists
	}
	return err
}

func (r *Repository) Set(key string, val []byte, item ...cache.ItemF) error {
	return r.put(true, key, val, nil, item...)
}

func (r *Repository) CompareAndSwap(
	key string,
	handler func(value *cache.Item) (*cache.Item, error),
	sleepDur ...time.Duration,
) error {
	for i := 0; i < casRetryLimit; i++ {
		data, generation, err := r.get(key)
		if err != nil {
			return err
		}

		newData, err := handler(data)
		if err != nil || newData == nil {
			return err
		}

		itemFs := make([]cache.ItemF, 0)
		if newData.Flags != 0 {
			itemFs = append(itemFs, cache.WithFlags(newData.Flags))
		}
		err = r.put(true, key, newData.Value, &generation, itemFs...)
		var asErr *aero.AerospikeError
		if errors.As(err, &asErr) && asErr.ResultCode == aeroTypes.GENERATION_ERROR {
			time.Sleep(casSleepTime * time.Millisecond)
			continue
		}
		if err != nil {
			return err
		}

		return nil
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

func (r *Repository) put(replace bool, key string, val []byte, generation *uint32, itemFs ...cache.ItemF) error {
	bop, batchWrites, err := r.prepBatchWrite(replace, key, val, generation, itemFs...)
	if err != nil {
		return err
	}
	return r.cache.BatchOperate(bop, batchWrites)
}

func (r *Repository) prepBatchWrite(replace bool, key string, val []byte, generation *uint32, itemFs ...cache.ItemF) (*aero.BatchPolicy, []aero.BatchRecordIfc, error) {
	asKey, err := r.key(key)
	if err != nil {
		return nil, nil, err
	}

	lenBin := aero.NewBin(lenBinKey, len(val))
	val, continuations := r.split(key, val)
	br := make([]aero.BatchRecordIfc, 0, len(continuations)+1)
	mainBins := []*aero.Bin{aero.NewBin(valueBinKey, val), lenBin}
	bwp := aero.NewBatchWritePolicy()
	bwp.Expiration = truncExpiration(r.liveTime)
	bwp.RecordExistsAction = aero.CREATE_ONLY
	if replace {
		bwp.RecordExistsAction = aero.REPLACE
	}
	if len(itemFs) > 0 {
		it := cache.CollectItem(itemFs...)
		if it.Flags != 0 {
			mainBins = append(mainBins, aero.NewBin(flagsBinKey, it.Flags))
		}
		if it.TTL != 0 {
			bwp.Expiration = truncExpiration(it.TTL)
		}
	}

	continuationKeys := make([]string, 0, len(continuations))
	for _, c := range continuations {
		asKey, err := r.key(c.Key)
		if err != nil {
			return nil, nil, err
		}
		br = append(br, aero.NewBatchWrite(bwp, asKey, aero.PutOp(aero.NewBin(valueBinKey, c.Value))))
		continuationKeys = append(continuationKeys, c.Key)
	}

	mainOps := make([]*aero.Operation, 0, len(mainBins))
	for _, b := range mainBins {
		mainOps = append(mainOps, aero.PutOp(b))
	}
	if len(continuationKeys) > 0 {
		mainOps = append(mainOps, aero.PutOp(aero.NewBin(continuationBinKey, continuationKeys)))
	}
	if generation != nil {
		bwp.GenerationPolicy = aero.EXPECT_GEN_EQUAL
		bwp.Generation = *generation
	}
	br = append(br, aero.NewBatchWrite(bwp, asKey, mainOps...))

	return r.cache.getBatchWritePolicy(), br, nil
}

func (r *Repository) split(key string, val []byte) ([]byte, []continuation) {
	if len(val) <= r.itemSizeLimit {
		return val, nil
	}

	estimatedChunks := (len(val) + r.itemSizeLimit - 1) / r.itemSizeLimit
	log.TraceWithFields(log.Fields{
		"key":    key,
		"size":   len(val),
		"chunks": estimatedChunks,
	}, "aerospike: value exceeds limit")

	continuations := make([]continuation, 0, estimatedChunks-1)
	for i := r.itemSizeLimit; i < len(val); i += r.itemSizeLimit {
		continuations = append(continuations, continuation{
			Key:   fmt.Sprintf("%s_%d", key, i/r.itemSizeLimit),
			Value: val[i : i+min(r.itemSizeLimit, len(val)-i)],
		})
	}
	return val[:r.itemSizeLimit], continuations
}

func truncExpiration(d time.Duration) uint32 {
	const maxDuration = 30 * 24 * time.Hour
	if d > maxDuration {
		return uint32(maxDuration.Seconds())
	}
	return uint32(d.Seconds())
}

type continuation struct {
	Key   string
	Value []byte
}

var _ cache.IDistrCacheRepository = (*Repository)(nil)
