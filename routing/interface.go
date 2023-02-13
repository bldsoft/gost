package routing

import (
	"net/http"
)

type Condition interface {
	Match(r *http.Request) (matched bool, err error)
}

type Action interface {
	Apply(http.Handler) http.Handler
}

type Rule interface {
	Condition
	Action
}
