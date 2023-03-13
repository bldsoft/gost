package routing

import (
	"errors"
	"net/http"
)

type IRule interface {
	Name() string
	Condition
	Action
}

type outgoingMatchFunc = func(w http.ResponseWriter, r *http.Request) (matched bool, err error)

type Condition interface {
	// outgoingMatch != nil means outgoing match is needed
	IncomingMatch(w http.ResponseWriter, r *http.Request) (matched bool, outgoingMatch outgoingMatchFunc, err error)
}

type actionFunc = func(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request, error)

var ErrStopHandling = errors.New("stop handling")

type Action interface {
	// return ErrStopHandling to stop request handling (e.g. for incoming redirect)
	DoBeforeHandle(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request, error)
	DoAfterHandle(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request, error)
}

type ValueMatcher[T any] interface {
	MatchValue(val T) (bool, error)
}

type ValueExtractor[T any] interface {
	IsIncoming() bool
	ExtractValue(w http.ResponseWriter, r *http.Request) T
}
