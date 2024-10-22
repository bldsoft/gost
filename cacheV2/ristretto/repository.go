package ristretto

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/bldsoft/gost/cache"
	"github.com/bldsoft/gost/log"
	"github.com/dgraph-io/ristretto"
)

var ErrorCacheSet = errors.New("set failed")

type Repository struct {
	cache *ristretto.Cache
}

func NewRepository(jsonConfig string) *Repository {
	defConfig := &ristretto.Config{
		NumCounters: 1e6, // number of keys to track frequency of (1M).
		MaxCost:     1e5, // maximum capacity (100K items).
		BufferItems: 64,  // number of keys per Get buffer.
	}
	if err := json.Unmarshal([]byte(jsonConfig), &defConfig); err != nil {
		log.WarnWithFields(log.Fields{"err": err}, "Failed to unmarshal Ristretto config")
	}
	client, err := ristretto.NewCache(defConfig)
	if err != nil {
		log.Panicf("Ristretto cache failed: %v", err)
	}
	return &Repository{cache: client}
}

func (r *Repository) Get(key string) ([]byte, error) {
	val, ok := r.cache.Get(key)
	if !ok {
		return nil, cache.ErrCacheMiss
	}
	return (val).([]byte), nil
}

func (r *Repository) Set(key string, value []byte) error {
	if ok := r.cache.Set(key, value, 1); ok {
		return nil
	}
	return ErrorCacheSet
}

func (r *Repository) SetFor(key string, value []byte, ttl time.Duration) error {
	if ok := r.cache.SetWithTTL(key, value, 1, ttl); ok {
		return nil
	}
	return ErrorCacheSet
}

func (r *Repository) Delete(key string) error {
	r.cache.Del(key)
	return nil
}

func (r *Repository) Reset() {
	r.cache.Clear()
}
