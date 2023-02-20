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

type ValueMatcher[T any] interface {
	MatchValue(val T) (bool, error)
}

type ValueExtractor[T any] interface {
	ExtractValue(r *http.Request) T
}
