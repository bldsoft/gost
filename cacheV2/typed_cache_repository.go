package cache2

import (
	"bytes"
	"encoding/gob"
	"time"
)

type TypedRepository[T any] struct {
	IExpiringCacheRepository
}

func Typed[T any](rep IExpiringCacheRepository) *TypedRepository[T] {
	return &TypedRepository[T]{rep}
}

func (r *TypedRepository[T]) Get(key string) (res T, err error) {
	data, err := r.IExpiringCacheRepository.Get(key)
	if err != nil {
		return res, err
	}
	return r.cacheUnmarshal(data)
}

func (r *TypedRepository[T]) Set(key string, value T) error {
	data, err := r.cacheMarshal(value)
	if err != nil {
		return err
	}
	return r.IExpiringCacheRepository.Set(key, data)
}

func (r *TypedRepository[T]) SetFor(key string, value T, ttl time.Duration) error {
	data, err := r.cacheMarshal(value)
	if err != nil {
		return err
	}
	return r.IExpiringCacheRepository.SetFor(key, data, ttl)
}

func (r *TypedRepository[T]) Delete(key string) error {
	return r.IExpiringCacheRepository.Delete(key)
}

func (r *TypedRepository[T]) Reset() {
	r.IExpiringCacheRepository.Reset()
}

func (r *TypedRepository[T]) cacheMarshal(e T) ([]byte, error) {
	if data, ok := any(e).([]byte); ok {
		return data, nil
	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(e); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (h *TypedRepository[T]) cacheUnmarshal(data []byte) (e T, err error) {
	if _, ok := any(e).([]byte); ok {
		return any(data).(T), nil
	}
	dec := gob.NewDecoder(bytes.NewBuffer(data))
	err = dec.Decode(&e)
	return
}
