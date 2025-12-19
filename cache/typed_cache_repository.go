package cache

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"time"
)

type Marshaler[T any] interface {
	Marshal(v T) ([]byte, error)
	Unmarshal(data []byte) (T, error)
}

type TypedRepository[T any] struct {
	IExpiringCacheRepository
	marshaler Marshaler[T]
}

func Typed[T any](rep IExpiringCacheRepository) *TypedRepository[T] {
	return &TypedRepository[T]{
		IExpiringCacheRepository: rep,
		marshaler:                GobMarshaler[T]{},
	}
}

func (r *TypedRepository[T]) WithMarshaler(marshaler Marshaler[T]) *TypedRepository[T] {
	return &TypedRepository[T]{
		IExpiringCacheRepository: r.IExpiringCacheRepository,
		marshaler:                marshaler,
	}
}

func (r *TypedRepository[T]) WithJSONMarshaler() *TypedRepository[T] {
	return r.WithMarshaler(JSONMarshaler[T]{})
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
	return r.marshaler.Marshal(e)
}

func (r *TypedRepository[T]) cacheUnmarshal(data []byte) (e T, err error) {
	if _, ok := any(e).([]byte); ok {
		return any(data).(T), nil
	}
	return r.marshaler.Unmarshal(data)
}

type GobMarshaler[T any] struct{}

func (h GobMarshaler[T]) Marshal(v T) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (h GobMarshaler[T]) Unmarshal(data []byte) (e T, err error) {
	dec := gob.NewDecoder(bytes.NewBuffer(data))
	err = dec.Decode(&e)
	return e, err
}

type JSONMarshaler[T any] struct{}

func (h JSONMarshaler[T]) Marshal(v T) ([]byte, error) {
	return json.Marshal(v)
}

func (h JSONMarshaler[T]) Unmarshal(data []byte) (e T, err error) {
	err = json.Unmarshal(data, &e)
	return e, err
}
