package feature

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Same as String feature.
// TODO: rewrite with generics when go 1.18 is released
type onBoolChangeHandler = func(bool)
type boolValidator = func(bool) error

// Bool ...
type Bool struct {
	ID               IdType
	value            bool
	validator        boolValidator
	onchangeHandlers []onBoolChangeHandler
}

func (f *Bool) GetID() IdType {
	return f.ID
}

func (f *Bool) onChange(newValue bool) {
	for _, handler := range f.onchangeHandlers {
		handler(newValue)
	}
}

// AddOnChangeHandler adds a handler for changing feature
func (f *Bool) AddOnChangeHandler(handler onBoolChangeHandler) {
	f.onchangeHandlers = append(f.onchangeHandlers, handler)
}

func (f *Bool) Get() bool {
	return f.value
}

// Set feature value
func (f *Bool) Set(value bool) error {
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

func (f *Bool) SetFromString(value string) error {
	val, err := strconv.ParseBool(value)
	if err != nil {
		return err
	}
	return f.Set(val)
}

func (f *Bool) Validate(value string) error {
	val, err := strconv.ParseBool(value)
	if err != nil {
		return err
	}
	return f.validate(val)
}

// Validate validates value for feature
func (f *Bool) validate(value bool) error {
	if f.validator == nil {
		return nil
	}
	return f.validator(value)
}

func (f *Bool) String() string {
	return strconv.FormatBool(f.value)
}

func (f *Bool) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.value)
}
