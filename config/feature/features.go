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

// NewString creates a new feature and put it to feature.Features
func NewString(id IdType, value string, validator func(string) error, handlers ...onStringChangeHandler) *String {
	feature := &String{ID: id, value: value, validator: validator, onchangeHandlers: handlers}
	Features.features[id] = feature
	return feature
}

// NewInt creates a new feature and put it to feature.Features
func NewInt(id IdType, value int, validator intValidator, handlers ...onIntChangeHandler) *Int {
	feature := &Int{ID: id, value: value, validator: validator, onchangeHandlers: handlers}
	Features.features[id] = feature
	return feature
}

// NewBool creates a new feature and put it to feature.Features
func NewBool(id IdType, value bool, validator boolValidator, handlers ...onBoolChangeHandler) *Bool {
	feature := &Bool{ID: id, value: value, validator: validator, onchangeHandlers: handlers}
	Features.features[id] = feature
	return feature
}
