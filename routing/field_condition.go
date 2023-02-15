package routing

import (
	"encoding/json"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
)

type FieldCondition struct {
	ValueExtractor
	ValueMatcher
}

func NewFieldCondition(e ValueExtractor, m ValueMatcher) Condition {
	return &FieldCondition{
		ValueExtractor: e,
		ValueMatcher:   m,
	}
}

func (f FieldCondition) Match(r *http.Request) (matched bool, err error) {
	v := f.ValueExtractor.ExtractValue(r)
	return f.ValueMatcher.MatchValue(v)
}

func (f *FieldCondition) UnmarshalJSON(b []byte) (err error) {
	return f.unmarshalHelper(b, json.Unmarshal)
}

func (f FieldCondition) MarshalJSON() ([]byte, error) {
	return f.marshalHelper(json.Marshal)
}

func (f *FieldCondition) UnmarshalBSON(b []byte) (err error) {
	return f.unmarshalHelper(b, bson.Unmarshal)
}

func (f FieldCondition) MarshalBSON() ([]byte, error) {
	return f.marshalHelper(bson.Marshal)
}

func (f FieldCondition) marshalHelper(marshalFunc func(val interface{}) ([]byte, error)) ([]byte, error) {
	fieldName, err := valueExtractorMarshaller.NameByValue(f.ValueExtractor)
	if err != nil {
		return nil, err
	}
	return marshalFunc(struct {
		Field string                         `json:"field" bson:"field"`
		Op    marshallingField[ValueMatcher] `json:"op" bson:"op"`
	}{
		Field: fieldName,
		Op:    marshallingField[ValueMatcher]{f.ValueMatcher, valueMatcherMarshaller},
	})
}

func (f *FieldCondition) unmarshalHelper(b []byte, unmarshalFunc func(data []byte, v any) error) (err error) {
	type outFieldCondition struct {
		Field string                         `json:"field" bson:"field"`
		Op    marshallingField[ValueMatcher] `json:"op" bson:"op"`
	}
	temp := &outFieldCondition{
		Op: marshallingField[ValueMatcher]{f.ValueMatcher, valueMatcherMarshaller},
	}
	if err := unmarshalFunc(b, &temp); err != nil {
		return err
	}
	f.ValueMatcher = temp.Op.Value
	f.ValueExtractor, err = valueExtractorMarshaller.AllocValue(temp.Field)
	if err != nil {
		return
	}
	return nil
}
