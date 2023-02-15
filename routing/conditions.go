package routing

import "net/http"

type NoCondition struct{}

func (c NoCondition) Match(r *http.Request) (matched bool, err error) {
	return true, nil
}

var NoCond = NoCondition{}
