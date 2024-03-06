package feature

import (
	"encoding/json"
	"errors"
	"fmt"
)

var ErrDisabled = errors.New("feature is disabled; read-only")

type Feature[T comparable] struct {
	ID    IdType
	value T

	depricated       bool
	parse            func(string) (T, error)
	validators       []func(T) error
	onchangeHandlers []func(T)
}

func (f *Feature[T]) GetID() IdType {
	return f.ID
}

func (f *Feature[T]) onChange(newValue T) {
	for _, handler := range f.onchangeHandlers {
		handler(newValue)
	}
}

func (f *Feature[T]) AddOnChangeHandler(handler func(T), handlers ...func(T)) *Feature[T] {
	f.onchangeHandlers = append(f.onchangeHandlers, handler)
	f.onchangeHandlers = append(f.onchangeHandlers, handlers...)
	return f
}

func (f *Feature[T]) SetValidator(validate func(T) error) *Feature[T] {
	f.validators = append(f.validators, validate)
	return f
}

// Get returns feature value
func (f *Feature[T]) Get() T {
	if f.depricated {
		panic(ErrDisabled)
	}
	return f.value
}

func (f *Feature[T]) validate(value T) error {
	if len(f.validators) == 0 {
		return nil
	}
	for _, v := range f.validators {
		if err := v(value); err != nil {
			return err
		}
	}
	return nil
}

func (f *Feature[T]) Validate(value string) error {
	val, err := f.parse(value)
	if err != nil {
		return err
	}
	return f.validate(val)
}

func (f *Feature[T]) Set(value T) error {
	if f.value == value {
		return nil
	}
	if err := f.validate(value); err != nil {
		return fmt.Errorf("not valid feature value: %w", err)
	}
	f.value = value
	f.onChange(value)
	return nil
}

func (f *Feature[T]) SetFromString(value string) error {
	val, err := f.parse(value)
	if err != nil {
		return err
	}
	return f.Set(val)
}

func (f *Feature[T]) String() string {
	return fmt.Sprintf("%v", f.value)
}

func (f *Feature[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.value)
}
