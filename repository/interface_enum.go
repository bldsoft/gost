// Code generated by go-enum DO NOT EDIT.
// Version:
// Revision:
// Build Date:
// Built By:

package repository

import (
 "errors"
 "fmt"
)

const (
 // EventTypeCreate is a EventType of type create.
 EventTypeCreate EventType = "create"
 // EventTypeUpdate is a EventType of type update.
 EventTypeUpdate EventType = "update"
 // EventTypeDelete is a EventType of type delete.
 EventTypeDelete EventType = "delete"
)

var ErrInvalidEventType = errors.New("not a valid EventType")

// String implements the Stringer interface.
func (x EventType) String() string {
 return string(x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x EventType) IsValid() bool {
 _, err := ParseEventType(string(x))
 return err == nil
}

var _EventTypeValue = map[string]EventType{
 "create": EventTypeCreate,
 "update": EventTypeUpdate,
 "delete": EventTypeDelete,
}

// ParseEventType attempts to convert a string to a EventType.
func ParseEventType(name string) (EventType, error) {
 if x, ok := _EventTypeValue[name]; ok {
  return x, nil
 }
 return EventType(""), fmt.Errorf("%s is %w", name, ErrInvalidEventType)
}