package routing

import (
	"github.com/bldsoft/gost/utils"
)

type MatcherAnyOf struct {
	Args []interface{} `json:"args" bson:"args"`
}

func MatchesAnyOf(args ...interface{}) MatcherAnyOf {
	return MatcherAnyOf{Args: args}
}

func (m MatcherAnyOf) MatchValue(val interface{}) (bool, error) {
	return utils.IsIn(val, m.Args...), nil
}
