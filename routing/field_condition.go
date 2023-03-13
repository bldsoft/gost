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

func (f FieldCondition[T]) IsIncoming() bool {
	return f.Extractor.IsIncoming()
}

func (f FieldCondition[T]) IncomingMatch(w http.ResponseWriter, r *http.Request) (matched bool, outgoingMatch outgoingMatchFunc, err error) {
	if f.Extractor.IsIncoming() {
		v := f.Extractor.ExtractValue(w, r)
		matched, err := f.Matcher.MatchValue(v)
		return matched, nil, err
	}

	return false, func(w http.ResponseWriter, r *http.Request) (matched bool, err error) {
		v := f.Extractor.ExtractValue(w, r)
		return f.Matcher.MatchValue(v)
	}, nil
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
	return marshalFunc(struct {
		Extractor marshallingField[ValueExtractor[T]] `json:"extractor,omitempty" bson:"extractor,omitempty"`
		Op        marshallingField[ValueMatcher[T]]   `json:"op" bson:"op"`
	}{
		Extractor: marshallingField[ValueExtractor[T]]{f.Extractor, valueExtractorMarshaller},
		Op:        marshallingField[ValueMatcher[T]]{f.Matcher, valueMatcherMarshaller},
	})
}

type outFieldCondition[T any] struct {
	Extractor marshallingField[ValueExtractor[T]] `json:"extractor,omitempty" bson:"extractor,omitempty"`
	Op        marshallingField[ValueMatcher[T]]   `json:"op" bson:"op"`
}

func (f *FieldCondition[T]) unmarshalHelper(b []byte, unmarshalFunc func(data []byte, v any) error) (err error) {
	valueExtractorMarshaller, valueMatcherMarshaller, err := f.marshallers()
	if err != nil {
		return err
	}

	temp := &outFieldCondition[T]{
		Op:        marshallingField[ValueMatcher[T]]{f.Matcher, valueMatcherMarshaller},
		Extractor: marshallingField[ValueExtractor[T]]{f.Extractor, valueExtractorMarshaller},
	}
	if err := unmarshalFunc(b, &temp); err != nil {
		return err
	}
	f.Matcher = temp.Op.Value
	f.Extractor = temp.Extractor.Value
	return nil
}
