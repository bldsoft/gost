package cache

import (
	"bytes"
	"encoding/gob"
	"time"
)

type expCacheEntry struct {
	Data []byte
	Exp  int64
}

type ExpiringCacheRepository struct {
	ILocalCacheRepository
}

func NewExpiringRepository(rep ILocalCacheRepository) *ExpiringCacheRepository {
	return &ExpiringCacheRepository{ILocalCacheRepository: rep}
}

func (r ExpiringCacheRepository) cacheMarshal(e expCacheEntry) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(e); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (h ExpiringCacheRepository) cacheUnmarshal(data []byte) (e expCacheEntry, err error) {
	dec := gob.NewDecoder(bytes.NewBuffer(data))
	err = dec.Decode(&e)
	return
}

func (r *ExpiringCacheRepository) set(key string, value []byte, exp int64) error {
	data, err := r.cacheMarshal(expCacheEntry{Data: value, Exp: exp})
	if err != nil {
		return err
	}
	return r.ILocalCacheRepository.Set(key, data)
}

func (r *ExpiringCacheRepository) SetFor(key string, value []byte, ttl time.Duration) error {
	return r.set(key, value, time.Now().Add(ttl).Unix())
}

func (r *ExpiringCacheRepository) Set(key string, value []byte) error {
	return r.set(key, value, 0)
}

func (r *ExpiringCacheRepository) Get(key string) ([]byte, error) {
	data, err := r.ILocalCacheRepository.Get(key)
	if err != nil {
		return nil, err
	}

	entry, err := r.cacheUnmarshal(data)
	if err != nil {
		return nil, err
	}
	if entry.Exp != 0 && time.Unix(entry.Exp, 0).Before(time.Now()) {
		return nil, ErrCacheMiss
	}

	return entry.Data, nil
}
