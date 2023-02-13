package routing

import "net/http"

type ActionFunc func(http.Handler) http.Handler

func (f ActionFunc) Apply(h http.Handler) http.Handler {
	return f(h)
}

var (
	HandleAction = ActionFunc(func(h http.Handler) http.Handler {
		return h
	})
)
