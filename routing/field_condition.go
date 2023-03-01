package routing

import (
	"encoding/json"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
)

const fieldConditionMatcherFieldName = "Matcher"
const fieldConditionExtractorFieldName = "Extractor"

type FieldCondition[T any] struct {
	Extractor ValueExtractor[T]
	Matcher   ValueMatcher[T]
}

func NewFieldCondition[T any](e ValueExtractor[T], m ValueMatcher[T]) Condition {
	return &FieldCondition[T]{
		Extractor: e,
		Matcher:   m,
	}
}

func (f FieldCondition[T]) Match(r *http.Request) (matched bool, err error) {
	v := f.Extractor.ExtractValue(r)
	return f.Matcher.MatchValue(v)
}

func (f *FieldCondition[T]) UnmarshalJSON(b []byte) (err error) {
	return f.unmarshalHelper(b, json.Unmarshal)
}

func (f FieldCondition[T]) MarshalJSON() ([]byte, error) {
	return f.marshalHelper(json.Marshal)
}

func (f *FieldCondition[T]) UnmarshalBSON(b []byte) (err error) {
	return f.unmarshalHelper(b, bson.Unmarshal)
}

func (f FieldCondition[T]) MarshalBSON() ([]byte, error) {
	return f.marshalHelper(bson.Marshal)
}

func (f FieldCondition[T]) marshallers() (*PolymorphMarshaller[ValueExtractor[T]], *PolymorphMarshaller[ValueMatcher[T]], error) {
	valueExtractorMarshaller, err := valueExtractorMarshaller[T]()
	if err != nil {
		return nil, nil, err
	}
	valueMatcherMarshaller, err := matcherMarshaller[T]()
	if err != nil {
		return nil, nil, err
	}
	return valueExtractorMarshaller, valueMatcherMarshaller, nil
}

func (f FieldCondition[T]) marshalHelper(marshalFunc func(val interface{}) ([]byte, error)) ([]byte, error) {
	valueExtractorMarshaller, valueMatcherMarshaller, err := f.marshallers()
	if err != nil {
		return nil, err
	}

	fieldName, err := valueExtractorMarshaller.NameByValue(f.Extractor)
	if err != nil {
		return nil, err
	}

	return marshalFunc(struct {
		Field string                            `json:"field" bson:"field"`
		Op    marshallingField[ValueMatcher[T]] `json:"op" bson:"op"`
	}{
		Field: fieldName,
		Op:    marshallingField[ValueMatcher[T]]{f.Matcher, valueMatcherMarshaller},
	})
}

func (f *FieldCondition[T]) unmarshalHelper(b []byte, unmarshalFunc func(data []byte, v any) error) (err error) {
	valueExtractorMarshaller, valueMatcherMarshaller, err := f.marshallers()
	if err != nil {
		return err
	}

	type outFieldCondition struct {
		Field string                            `json:"field" bson:"field"`
		Op    marshallingField[ValueMatcher[T]] `json:"op" bson:"op"`
	}
	temp := &outFieldCondition{
		Op: marshallingField[ValueMatcher[T]]{f.Matcher, valueMatcherMarshaller},
	}
	if err := unmarshalFunc(b, &temp); err != nil {
		return err
	}
	f.Matcher = temp.Op.Value
	f.Extractor, err = valueExtractorMarshaller.AllocValue(temp.Field)
	if err != nil {
		return
	}
	return nil
}
