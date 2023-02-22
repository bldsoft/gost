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

const ArgsMethodName = "Args"

type IArgs[A any] interface {
	SetArgs(A) error
	Args() A
}

type ValueMatcher[T any, A any] interface {
	IArgs[A]
	MatchValue(val T) (bool, error)
}

type ValueExtractor[T any] interface {
	ExtractValue(r *http.Request) T
}
