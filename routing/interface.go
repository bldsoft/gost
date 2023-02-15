package routing

import (
	"net/http"
)

type IRule interface {
	Condition
	Action
}

type Condition interface {
	Match(r *http.Request) (matched bool, err error)
}

type Action interface {
	Apply(http.Handler) http.Handler
}

type ValueMatcher interface {
	MatchValue(val interface{}) (bool, error)
}

type ValueExtractor interface {
	ExtractValue(r *http.Request) interface{}
}
