package routing

import (
	"encoding/json"
	"net/http"
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
	type outFieldCondition struct {
		Field string                         `json:"field"`
		Op    marshallingField[ValueMatcher] `json:"op"`
	}
	temp := &outFieldCondition{
		Op: marshallingField[ValueMatcher]{f.ValueMatcher, valueMatcherMarshaller},
	}
	if err := json.Unmarshal(b, &temp); err != nil {
		return err
	}
	f.ValueExtractor, err = valueExtractorMarshaller.AllocValue(temp.Field)
	if err != nil {
		return
	}
	f.ValueMatcher = temp.Op.Value
	return nil
}

func (f FieldCondition) MarshalJSON() ([]byte, error) {
	fieldName, err := valueExtractorMarshaller.NameByValue(f.ValueExtractor)
	if err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		Field string                         `json:"field"`
		Op    marshallingField[ValueMatcher] `json:"op"`
	}{
		Field: fieldName,
		Op:    marshallingField[ValueMatcher]{f.ValueMatcher, valueMatcherMarshaller},
	})
}
