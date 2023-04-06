package cache

import (
	"errors"
	"time"
)

var (
	ErrCacheMiss = errors.New("cache miss")
)

// ILocalCacheRepository ...
type ILocalCacheRepository[T any] interface {
	Get(key string) (T, error)
	Set(key string, value T) error
	Delete(key string) error
	Reset()
}

type IExpiringCacheRepository[T any] interface {
	ILocalCacheRepository[T]
	SetFor(key string, value T, ttl time.Duration) error
}

// IDistrCacheRepository ...
type IDistrCacheRepository[T any] interface {
	IExpiringCacheRepository[T]

	//TODO: add it to ILocalCacheRepository
	GetWithFlags(key string) (data T, flags uint32, err error)
	Exist(key string) bool
	Add(key string, value T) error
	AddFor(key string, value T, ttl time.Duration) error
	SetWithFlags(key string, value T, flags uint32) error
	CompareAndSwap(key string, handler func(value T) (T, error)) error
}
