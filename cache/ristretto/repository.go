package ristretto

import (
	"encoding/json"
	"errors"

	"github.com/bldsoft/gost/log"
	"github.com/dgraph-io/ristretto"
)

var ErrorCacheSet = errors.New("set failed")
var ErrorNotFound = errors.New("not found")
var ErrorWrongType = errors.New("wrong type")

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
		return nil, ErrorNotFound
	}
	bytes, ok := (val).([]byte)
	if !ok {
		return nil, ErrorWrongType
	}
	return bytes, nil
}

func (r *Repository) Set(key string, value []byte) error {
	ok := r.cache.Set(key, value, 1)
	if ok {
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
