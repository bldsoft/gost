package feature

import (
	"time"

	"github.com/bldsoft/gost/utils"
)

type Duration struct {
	*Feature[time.Duration]
	multiplier time.Duration
}

// NewDuration creates feature with time.Duration type.
// The init feature value is dur * multiplier
func NewDuration(id IdType, dur, multiplier time.Duration) *Duration {
	return &Duration{Feature: NewFeature(id, time.Duration(dur)*multiplier), multiplier: multiplier}
}

// NewSeconds create a new duration feature with time.Second multiplier.
// It means that database must store a number of seconds.
// The values that will be passed to the handlers will be in seconds.
// No multiplication is nessesary (like Feature.Get() * time.Second), just use the values.
func NewSeconds(id IdType, sec int64) *Duration {
	return NewDuration(id, time.Duration(sec), time.Second)
}

// NewSeconds create a new duration feature with time.Minute multiplier.
// Same as NewSeconds but for minutes.
func NewMinutes(id IdType, mins int64) *Duration {
	return NewDuration(id, time.Duration(mins), time.Minute)
}

// AddOnChangeHandler adds a handler to be called when feature changes. The method returns the same feature
func (f *Duration) AddOnChangeHandler(handler func(time.Duration)) *Duration {
	f.Feature.AddOnChangeHandler(handler)
	return f
}

// SetValidator adds validate function. The method returns the same feature
func (f *Duration) SetValidator(validate func(time.Duration) error) *Duration {
	f.Feature.SetValidator(validate)
	return f
}

func (f *Duration) SetFromString(value string) error {
	val, err := utils.Parse[time.Duration](value)
	if err != nil {
		return err
	}
	return f.Feature.Set(val * f.multiplier)
}
