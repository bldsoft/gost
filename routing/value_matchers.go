package routing

import (
	"github.com/bldsoft/gost/utils"
)

type MatcherAnyOf[T comparable] struct {
	Values []T `json:"args,omitempty" bson:"args,omtempty"`
}

func MatchesAnyOf[T comparable](args ...T) *MatcherAnyOf[T] {
	return &MatcherAnyOf[T]{Values: args}
}

func (m MatcherAnyOf[T]) Args() []T {
	return m.Values
}

func (m *MatcherAnyOf[T]) SetArgs(args []T) error {
	m.Values = args
	return nil
}

func (m MatcherAnyOf[T]) MatchValue(val T) (bool, error) {
	return utils.IsIn(val, m.Values...), nil
}
