package feature

import (
	"errors"
	"fmt"
	"strconv"
)

// Same as String feature.
// TODO: rewrite with generics when go 1.18 is released
type onIntChangeHandler = func(int)
type intValidator = func(int) error

// Int ...
type Int struct {
	ID               IdType
	value            int
	validator        intValidator
	onchangeHandlers []onIntChangeHandler
}

// NewFeature creates new Feature
func NewInt(id IdType, value int, validator intValidator, handlers ...onIntChangeHandler) *Int {
	feature := &Int{ID: id, value: value, validator: validator, onchangeHandlers: handlers}
	Features.features[id] = feature
	return feature
}

func (f *Int) GetID() IdType {
	return f.ID
}

func (f *Int) onChange(newValue int) {
	for _, handler := range f.onchangeHandlers {
		handler(newValue)
	}
}

// AddOnChangeHandler adds handler for changing feature
func (f *Int) AddOnChangeHandler(handler onIntChangeHandler) {
	f.onchangeHandlers = append(f.onchangeHandlers, handler)
}

func (f *Int) Get() int {
	return f.value
}

// Set feature value
func (f *Int) Set(value int) error {
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

func (f *Int) SetFromString(value string) error {
	val, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	return f.Set(val)
}

func (f *Int) Validate(value string) error {
	val, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	return f.validate(val)
}

// Validate validates value for feature
func (f *Int) validate(value int) error {
	if f.validator == nil {
		return nil
	}
	return f.validator(value)
}

func (f *Int) String() string {
	return strconv.Itoa(f.value)
}

func ValidateUint(v int) error {
	if v < 0 {
		return errors.New("Negative")
	}
	return nil
}
