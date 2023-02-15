package routing

import "net/http"

type HandleAction struct{}

func (f HandleAction) Apply(h http.Handler) http.Handler {
	return h
}

var HandleAct = HandleAction{}
