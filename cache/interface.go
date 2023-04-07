package cache

import (
	"errors"
	"time"
)

var (
	ErrCacheMiss = errors.New("cache miss")
)

// ILocalCacheRepository ...
type ILocalCacheRepository interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte) error
	Delete(key string) error
	Reset()
}

type IExpiringCacheRepository interface {
	ILocalCacheRepository
	SetFor(key string, value []byte, ttl time.Duration) error
}

// IDistrCacheRepository ...
type IDistrCacheRepository interface {
	IExpiringCacheRepository

	//TODO: add it to ILocalCacheRepository
	GetWithFlags(key string) (data []byte, flags uint32, err error)
	Exist(key string) bool
	Add(key string, value []byte) error
	AddFor(key string, value []byte, ttl time.Duration) error
	SetWithFlags(key string, value []byte, flags uint32) error
	CompareAndSwap(key string, handler func(value []byte) ([]byte, error)) error
}

type Repository[T any] interface {
	Get(key string) (T, error)
	Set(key string, value T) error
	SetFor(key string, value T, ttl time.Duration) error
	Delete(key string) error
	Reset()
}
