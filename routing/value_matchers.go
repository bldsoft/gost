package routing

import "github.com/bldsoft/gost/utils"

type MatcherAnyOf[T comparable] struct {
	Args []T
}

func MatchesAnyOf[T comparable](args ...T) MatcherAnyOf[T] {
	return MatcherAnyOf[T]{Args: args}
}

func (m MatcherAnyOf[T]) MatchValue(val T) (bool, error) {
	return utils.IsIn(val, m.Args...), nil
}
