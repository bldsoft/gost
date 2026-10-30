package aerospike

import (
	"cmp"
	"crypto/rand"
	"errors"
	"fmt"
	"slices"
	"time"

	aero "github.com/aerospike/aerospike-client-go/v8"
	aeroTypes "github.com/aerospike/aerospike-client-go/v8/types"

	"github.com/bldsoft/gost/cache/v2"
	"github.com/bldsoft/gost/log"
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
	res, _, _, err := r.get(key)

	return res, err
}

func (r *Repository) get(key string) (*cache.Item, uint32, []string, error) {
	asKey, err := r.key(key)
	if err != nil {
		return nil, 0, nil, err
	}
	item, err := r.cache.Get(r.cache.getReadPolicy(), asKey)
	if err != nil || item == nil {
		if errors.Is(err, aero.ErrKeyNotFound) {
			return nil, 0, nil, cache.ErrCacheMiss
		}

		return nil, 0, nil, err
	}

	res := &cache.Item{}
	res.TTL = time.Duration(item.Expiration) * time.Second
	if flags, ok := item.Bins[flagsBinKey]; ok {
		res.Flags = uint32(flags.(int))
	}

	mainValue := item.Bins[valueBinKey].([]byte)
	totalSize := len(mainValue)

	if item.Bins[continuationBinKey] == nil {
		res.Value = mainValue

		return res, item.Generation, nil, nil
	}

	contKeyStrs := r.continuationKeysFromBin(item.Bins[continuationBinKey])
	continuationKeys := make([]*aero.Key, len(contKeyStrs))
	for i, k := range contKeyStrs {
		asKey, err = r.key(k)
		if err != nil {
			return nil, 0, nil, err
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

	if l, ok := item.Bins[lenBinKey]; ok {
		totalSize = l.(int)
	}
	res.Value = make([]byte, 0, totalSize)
	res.Value = append(res.Value, mainValue...)

	if len(continuationKeys) > 0 {
		continuations, err := r.cache.BatchGet(nil, continuationKeys, valueBinKey)
		if err != nil {
			return nil, 0, nil, err
		}
		for i, c := range continuations {
			if c == nil {
				return nil, 0, nil, fmt.Errorf("%w: aerospike: missing continuation record #%v for %s", cache.ErrCacheMiss, i, key)
			}
			if c.Bins == nil {
				return nil, 0, nil, fmt.Errorf("%w: aerospike: invalid continuation record #%v for %s (no bins)", cache.ErrCacheMiss, i, key)
			}
			val, ok := c.Bins[valueBinKey].([]byte)
			if !ok {
				return nil, 0, nil, fmt.Errorf("%w: aerospike: invalid value type in continuation record #%v for %s", cache.ErrCacheMiss, i, key)
			}
			res.Value = append(res.Value, val...)
		}
	}

	return res, item.Generation, contKeyStrs, nil
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

	prev, err := r.readContinuationKeys(key)
	if err != nil {
		return err
	}

	_, err = r.cache.Delete(nil, asKey)
	if err != nil {
		return err
	}

	r.deleteContinuations(prev)

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
	prev, err := r.readContinuationKeys(key)
	if err != nil {
		return err
	}

	if err = r.put(true, key, val, nil, item...); err != nil {
		return err
	}
	r.deleteContinuations(prev)

	return nil
}

func (r *Repository) CompareAndSwap(
	key string,
	handler func(value *cache.Item) (*cache.Item, error),
	sleepDur ...time.Duration,
) error {
	for range casRetryLimit {
		data, generation, prev, err := r.get(key)
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
		r.deleteContinuations(prev)

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

	if err = r.Add(key, val, opts...); errors.Is(err, cache.ErrExists) {
		i, err = r.Get(key)

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
	it := cache.CollectItem(itemFs...)
	if len(val) <= r.itemSizeLimit {
		return r.putRaw(replace, key, val, generation, it)
	}

	mainVal, continuations := r.split(key, rand.Text(), val)
	records, err := r.prepBatchWrite(replace, key, mainVal, continuations, len(val), generation, it)
	if err != nil {
		return err
	}
	if err = r.cache.BatchOperate(r.cache.getBatchWritePolicy(), records); err != nil {
		return err
	}

	return r.batchWriteErr(records)
}

func (r *Repository) putRaw(replace bool, key string, val []byte, generation *uint32, it *cache.Item) error {
	asKey, err := r.key(key)
	if err != nil {
		return err
	}
	wp := r.cache.getWritePolicy(0, r.expiration(it))
	wp.RecordExistsAction = aero.CREATE_ONLY
	if replace {
		wp.RecordExistsAction = aero.REPLACE
	}
	bins := aero.BinMap{valueBinKey: val}
	if it.Flags != 0 {
		bins[flagsBinKey] = it.Flags
	}
	if generation != nil {
		wp.GenerationPolicy = aero.EXPECT_GEN_EQUAL
		wp.Generation = *generation
	}

	return r.cache.Put(wp, asKey, bins)
}

func (r *Repository) batchWriteErr(records []aero.BatchRecordIfc) error {
	var errs []error
	last := len(records) - 1
	for i, rec := range slices.Backward(records) {
		br := rec.BatchRec()
		if br.ResultCode == aeroTypes.OK {
			continue
		}

		recErr := error(br.Err)
		if recErr == nil {
			recErr = &aero.AerospikeError{ResultCode: br.ResultCode}
		}
		if i != last {
			userKey := ""
			if br.Key != nil {
				userKey = br.Key.String()
			}
			recErr = fmt.Errorf("aerospike: continuation %s: %w", userKey, recErr)
		}
		errs = append(errs, recErr)
	}

	return errors.Join(errs...)
}

func (r *Repository) prepBatchWrite(replace bool, key string, val []byte, continuations []continuation, totalSize int, generation *uint32, it *cache.Item) ([]aero.BatchRecordIfc, error) {
	mainKey, err := r.key(key)
	if err != nil {
		return nil, err
	}

	expiration := r.expiration(it)
	br := make([]aero.BatchRecordIfc, 0, len(continuations)+1)

	contPolicy := aero.NewBatchWritePolicy()
	contPolicy.Expiration = expiration
	contPolicy.RecordExistsAction = aero.CREATE_ONLY

	continuationKeys := make([]string, 0, len(continuations))
	for _, c := range continuations {
		var contKey *aero.Key
		contKey, err = r.key(c.Key)
		if err != nil {
			return nil, err
		}
		br = append(br, aero.NewBatchWrite(contPolicy, contKey, aero.PutOp(aero.NewBin(valueBinKey, c.Value))))
		continuationKeys = append(continuationKeys, c.Key)
	}

	mainPolicy := aero.NewBatchWritePolicy()
	mainPolicy.Expiration = expiration
	mainPolicy.RecordExistsAction = aero.CREATE_ONLY
	if replace {
		mainPolicy.RecordExistsAction = aero.REPLACE
	}
	if generation != nil {
		mainPolicy.GenerationPolicy = aero.EXPECT_GEN_EQUAL
		mainPolicy.Generation = *generation
	}

	mainBins := []*aero.Bin{
		aero.NewBin(valueBinKey, val),
		aero.NewBin(lenBinKey, totalSize),
		aero.NewBin(continuationBinKey, continuationKeys),
	}
	if it.Flags != 0 {
		mainBins = append(mainBins, aero.NewBin(flagsBinKey, it.Flags))
	}
	mainOps := make([]*aero.Operation, len(mainBins))
	for i, b := range mainBins {
		mainOps[i] = aero.PutOp(b)
	}
	br = append(br, aero.NewBatchWrite(mainPolicy, mainKey, mainOps...))

	return br, nil
}

func (r *Repository) readContinuationKeys(key string) ([]string, error) {
	asKey, err := r.key(key)
	if err != nil {
		return nil, err
	}
	item, err := r.cache.Get(r.cache.getReadPolicy(), asKey, continuationBinKey)
	if err != nil {
		if errors.Is(err, aero.ErrKeyNotFound) {
			return nil, nil
		}

		return nil, err
	}
	if item == nil || item.Bins == nil {
		return nil, nil
	}

	return r.continuationKeysFromBin(item.Bins[continuationBinKey]), nil
}

func (r *Repository) deleteContinuations(keys []string) {
	batchDeletes := make([]aero.BatchRecordIfc, 0, len(keys))
	for _, k := range keys {
		asKey, err := r.key(k)
		if err != nil {
			log.WarnWithFields(log.Fields{"key": k, "err": err}, "failed to create key for continuation deletion")

			continue
		}
		batchDeletes = append(batchDeletes, aero.NewBatchDelete(nil, asKey))
	}
	if len(batchDeletes) == 0 {
		return
	}
	if err := r.cache.BatchOperate(nil, batchDeletes); err != nil {
		log.WarnWithFields(log.Fields{"err": err}, "failed to delete continuation records")
	}
}

func (r *Repository) expiration(it *cache.Item) uint32 {
	return truncExpiration(cmp.Or(it.TTL, r.liveTime))
}

func (r *Repository) split(key, nonce string, val []byte) ([]byte, []continuation) {
	estimatedChunks := (len(val) + r.itemSizeLimit - 1) / r.itemSizeLimit
	log.TraceWithFields(log.Fields{
		"key":    key,
		"size":   len(val),
		"chunks": estimatedChunks,
	}, "aerospike: value exceeds limit")

	continuations := make([]continuation, 0, estimatedChunks-1)
	for i := r.itemSizeLimit; i < len(val); i += r.itemSizeLimit {
		continuations = append(continuations, continuation{
			Key:   fmt.Sprintf("%s_%s_%d", key, nonce, i/r.itemSizeLimit),
			Value: val[i : i+min(r.itemSizeLimit, len(val)-i)],
		})
	}

	return val[:r.itemSizeLimit], continuations
}

func (r *Repository) continuationKeysFromBin(raw any) []string {
	if raw == nil {
		return nil
	}
	values := raw.([]any)
	keys := make([]string, len(values))
	for i, k := range values {
		keys[i] = k.(string)
	}

	return keys
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
