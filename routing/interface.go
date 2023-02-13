package routing

import (
	"net/http"

	"github.com/bldsoft/gost/log"
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

type rule struct {
	Condition
	Action
}

func NewRule(cond Condition, action Action) Rule {
	return rule{
		Condition: cond,
		Action:    action,
	}
}

func Routing(rule Rule) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			matched, err := rule.Match(r)
			switch {
			case matched:
				rule.Apply(next).ServeHTTP(w, r)
			case err != nil:
				log.FromContext(r.Context()).ErrorWithFields(log.Fields{"err": err}, "Routing: checking the rule condition for the request")
				next.ServeHTTP(w, r)
			default:
				next.ServeHTTP(w, r)
			}
		})
	}
}
