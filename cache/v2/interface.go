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

type IDistrCacheRepository interface {
	Get(key string) (*Item, error)
	Exist(key string) bool
	Set(key string, val []byte, opts ...ItemF) error
	Add(key string, val []byte, opts ...ItemF) error
	AddOrGet(key string, val []byte, opts ...ItemF) (i *Item, added bool, err error)
	Delete(key string) error
	CompareAndSwap(key string, handler func(value *Item) (*Item, error), sleepDur ...time.Duration) error
}
