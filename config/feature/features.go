package feature

import (
	"strconv"
	"time"

	"github.com/bldsoft/gost/utils"
)

type (
	Bool     = Feature[bool]
	Int      = Feature[int]
	String   = Feature[string]
	Duration = Feature[time.Duration]
)

type IdType = int

func IdFromString(str string) IdType {
	id, _ := strconv.Atoi(str)
	return id
}

type IFeature interface {
	GetID() IdType
	SetFromString(string) error
	Validate(string) error
	String() string
}

type featureConfig struct {
	features map[IdType]IFeature
}

// Features contain all the features and provide access to them by ID
var Features = featureConfig{make(map[IdType]IFeature)}

// Get returns Feature by id
func (fc *featureConfig) Get(featureID IdType) IFeature {
	feature, ok := fc.features[featureID]
	if !ok {
		return nil
	}
	return feature
}

// NewCustomFeature create a new feature from any comparable type.
// For basic types you can use NewFeature function.
func NewCustomFeature[T comparable](id IdType, value T, parse func(string) (T, error), depricated ...bool) *Feature[T] {
	feature := &Feature[T]{ID: id, value: value, parse: parse}
	if len(depricated) > 0 {
		feature.depricated = depricated[0]
		feature.SetValidator(func(T) error {
			return ErrDisabled
		})
	}
	Features.features[id] = feature
	return feature
}

// NewFeature creates a new feature and put it to feature.Features
func NewFeature[T utils.Parsed](id IdType, value T, depricated ...bool) *Feature[T] {
	return NewCustomFeature(id, value, utils.Parse[T], depricated...)
}

// NewDuration creates a new feature for time.Duration type.
// time.ParseDuration is used for value parsing, so you can set value such as "300ms" or "2h45m".
func NewDuration(id IdType, dur time.Duration, depricated ...bool) *Duration {
	return NewCustomFeature(id, dur, time.ParseDuration, depricated...)
}
