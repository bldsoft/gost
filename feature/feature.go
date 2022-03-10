package feature

import (
	"strconv"

	"github.com/bldsoft/gost/config/feature"
)

type serviceValue struct {
	SrvName string
	Value   string
}

// Feature ...
type Feature struct {
	ID          feature.IdType  `bson:"_id,omitempty" json:"id"`
	Name        string          `bson:"name,omitempty" json:"name"`
	Description *string         `bson:"description,omitempty" json:"description"`
	GlobalValue *string         `bson:"globalValue,omitempty" json:"globalValue"`
	SrvValues   *[]serviceValue `bson:"srvvalues,omitempty" json:"srvValues,omitempty"`
}

func NewFeature(feature feature.IFeature, name string, description string) *Feature {
	value := feature.String()
	return &Feature{ID: feature.GetID(), Name: name, Description: &description, GlobalValue: &value}
}

func (f *Feature) GetID() interface{} {
	return f.ID
}

func (f *Feature) GenerateID() {
	//ID must be set explicitly
}

func (f *Feature) SetIDFromString(id string) error {
	f.ID = feature.IdFromString(id)
	return nil
}

func (f *Feature) IsNilID() bool {
	return f.ID == 0
}

func (f *Feature) StringID() string {
	return strconv.Itoa(f.ID)
}
