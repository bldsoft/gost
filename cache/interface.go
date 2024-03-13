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

// IDistrCacheRepository ...
type IDistrCacheRepository interface {
	IExpiringCacheRepository

	// TODO: add it to ILocalCacheRepository
	GetWithFlags(key string) (data []byte, flags uint32, err error)
	Exist(key string) bool
	Add(key string, value []byte) error
	AddFor(key string, value []byte, ttl time.Duration) error
	AddWithFlags(key string, value []byte, flags uint32) error
	AddForWithFlags(key string, value []byte, flags uint32, ttl time.Duration) error
	SetWithFlags(key string, value []byte, flags uint32) error
	SetForWithFlags(key string, value []byte, flags uint32, ttl time.Duration) error
	CompareAndSwap(key string, handler func(value []byte) ([]byte, error)) error
}

type DistrCacheRepository interface {
	Get(key string) ([]byte, uint32, error)
	Exist(key string) bool
	Set(key string, opts *Item) error
	Add(key string, opts *Item) error
	Delete(key string) error
	CompareAndSwap(key string, handler func(value []byte) ([]byte, error)) error
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
