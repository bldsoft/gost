package cache

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
	Get(key string) ([]byte, uint32, error)
	Exist(key string) bool
	Set(key string, opts *Item) error
	Add(key string, opts *Item) error
	Delete(key string) error
	CompareAndSwap(key string, handler func(value *Item) (*Item, error), sleepDur ...time.Duration) error
}

type Item struct {
	Value []byte
	TTL   *time.Duration
	Flags *uint32
}

type Repository[T any] interface {
	Get(key string) (T, error)
	Set(key string, value T) error
	SetFor(key string, value T, ttl time.Duration) error
	Delete(key string) error
	Reset()
}
