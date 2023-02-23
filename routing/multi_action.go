package routing

import (
	"net/http"
)

type MultiAction struct {
	Actions []Action
}

func JoinActions(action ...Action) MultiAction {
	return MultiAction{
		Actions: action,
	}
}

func (a MultiAction) Apply(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, action := range a.Actions {
			h = action.Apply(h)
		}
		h.ServeHTTP(w, r)
	})
}
