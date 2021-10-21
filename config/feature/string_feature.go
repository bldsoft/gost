package feature

import (
	"encoding/json"
	"fmt"
)

type onStringChangeHandler = func(string)
type stringValidator = func(string) error

// String ...
type String struct {
	ID               IdType
	value            string
	validator        stringValidator
	onchangeHandlers []onStringChangeHandler
}

// NewFeature creates new Feature
func NewString(id IdType, value string, validator func(string) error, handlers ...onStringChangeHandler) *String {
	feature := &String{ID: id, value: value, validator: validator, onchangeHandlers: handlers}
	Features.features[id] = feature
	return feature
}

func (f *String) GetID() IdType {
	return f.ID
}

func (f *String) onChange(newValue string) {
	for _, handler := range f.onchangeHandlers {
		handler(newValue)
	}
}

// AddOnChangeHandler adds handler for changing feature
func (f *String) AddOnChangeHandler(handler func(string)) {
	f.onchangeHandlers = append(f.onchangeHandlers, handler)
}

func (f *String) Get() string {
	return f.value
}

// Set feature value
func (f *String) Set(value string) error {
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

func (f *String) SetFromString(value string) error {
	return f.Set(value)
}

func (f *String) Validate(value string) error {
	return f.validate(value)
}

// Validate validates value for feature
func (f *String) validate(value string) error {
	if f.validator == nil {
		return nil
	}
	return f.validator(value)
}

func (f *String) String() string {
	return f.value
}

func (f *String) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.value)
}
