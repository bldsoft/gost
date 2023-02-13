package routing

import "net/http"

type ConditionFunc func(r *http.Request) (matched bool, err error)

func (f ConditionFunc) Match(r *http.Request) (matched bool, err error) {
	return f(r)
}

var (
	NoCondition = ConditionFunc(func(r *http.Request) (matched bool, err error) { return true, nil })
)
