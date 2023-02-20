package routing

import (
	"github.com/bldsoft/gost/utils"
)

type MatcherAnyOf struct {
	Args []string `json:"args" bson:"args"`
}

func MatchesAnyOf(args ...string) MatcherAnyOf {
	return MatcherAnyOf{Args: args}
}

func (m MatcherAnyOf) MatchValue(val string) (bool, error) {
	return utils.IsIn(val, m.Args...), nil
}
