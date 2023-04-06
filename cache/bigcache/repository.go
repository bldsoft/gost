package bigcache

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/allegro/bigcache"
	"github.com/bldsoft/gost/log"
)

var (
	// ErrNotImplemented means repository method not implemented
	ErrNotImplemented = errors.New("BigCacheRepositor: method not implemented")
)

type Repository struct {
	cache *bigcache.BigCache
}

func NewRepository(jsonConfig string) *Repository {
	defConfig := bigcache.DefaultConfig(time.Minute)
	if err := json.Unmarshal([]byte(jsonConfig), &defConfig); err != nil {
		log.WarnWithFields(log.Fields{"err": err}, "Failed to unmarshal Ristretto config")
	}
	client, err := bigcache.NewBigCache(defConfig)
	if err != nil {
		log.Fatalf("BigCache failed: %v", err)
	}

	return &Repository{cache: client}
}

func (r *Repository) Add(key string, value []byte) error {
	return r.Set(key, value)
}

func (r *Repository) AddFor(key string, value []byte, expiration time.Duration) error {
	return r.Set(key, value)
}

func (r *Repository) Get(key string) ([]byte, error) {
	res, err := r.cache.Get(key)
	return res, err
}

func (r *Repository) Set(key string, value []byte) error {
	return r.cache.Set(key, value)
}

func (r *Repository) SetFor(key string, value []byte, expiration time.Duration) error {
	return r.cache.Set(key, value)
}

func (r *Repository) Delete(key string) error {
	return r.cache.Delete(key)
}

func (r *Repository) Reset() {
	r.cache.Reset()
}

func (r *Repository) CompareAndSwap(key string, handler func(value []byte) ([]byte, error)) error {
	return ErrNotImplemented
}

func (r *Repository) GetWithFlags(key string) (data []byte, flags uint32, err error) {
	return nil, 0, ErrNotImplemented
}

func (r *Repository) SetWithFlags(key string, value []byte, flags uint32) error {
	return ErrNotImplemented
}

func (r *Repository) Exist(key string) bool {
	_, err := r.Get(key)
	return err == nil
}
