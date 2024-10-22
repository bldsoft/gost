package cache2

import (
	"errors"
	"time"
)

var (
	ErrCacheMiss = errors.New("cache miss")
	ErrExists    = errors.New("already exists")
)

//go:generate go run github.com/vektra/mockery/v2 --all --with-expecter

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

type IDistrCacheRepository interface {
	Get(key string) (*Item, error)
	Exist(key string) bool
	Set(key string, val []byte, opts ...ItemF) error
	Add(key string, val []byte, opts ...ItemF) error
	AddOrGet(key string, val []byte, opts ...ItemF) (i *Item, added bool, err error)
	Delete(key string) error
	CompareAndSwap(key string, handler func(value *Item) (*Item, error), sleepDur ...time.Duration) error
}

type Repository[T any] interface {
	Get(key string) (T, error)
	Set(key string, value T) error
	SetFor(key string, value T, ttl time.Duration) error
	Delete(key string) error
	Reset()
}
