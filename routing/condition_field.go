package routing

import (
	"net/http"
)

type ValueExtractor[T any] interface {
	ExtractValue(r *http.Request) T
}

type ValueMatcher[T comparable] interface {
	MatchValue(val T) (bool, error)
}

type FieldCondition[T comparable] interface {
	ValueExtractor[T]
	ValueMatcher[T]
}

type fieldCondition[T comparable] struct {
	ValueExtractor[T]
	ValueMatcher[T]
}

func NewFieldCondition[T comparable](e ValueExtractor[T], m ValueMatcher[T]) Condition {
	return fieldCondition[T]{
		ValueExtractor: e,
		ValueMatcher:   m,
	}
}

func (f fieldCondition[T]) Match(r *http.Request) (matched bool, err error) {
	v := f.ValueExtractor.ExtractValue(r)
	return f.ValueMatcher.MatchValue(v)
}
