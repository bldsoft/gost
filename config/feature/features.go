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

// Features ...
var Features = featureConfig{make(map[IdType]IFeature)}

// Get ...
func (fc *featureConfig) Get(featureID IdType) IFeature {
	feature, ok := fc.features[featureID]
	if !ok {
		return nil
	}
	return feature
}
