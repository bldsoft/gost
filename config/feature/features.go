package feature

import "strconv"

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

// Get ...
func (fc *featureConfig) Get(featureID IdType) IFeature {
	feature, ok := fc.features[featureID]
	if !ok {
		return nil
	}
	return feature
}

// NewFeature creates a new feature and put it to feature.Features
func NewFeature[T FeatureType](id IdType, value T) *Feature[T] {
	feature := &Feature[T]{ID: id, value: value}
	Features.features[id] = feature
	return feature
}
