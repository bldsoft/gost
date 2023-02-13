package routing

import (
	"net/http"

	"github.com/bldsoft/gost/log"
)

type ruleList struct {
	rules []Rule
}

func JoinRules(rules ...Rule) Rule {
	return &ruleList{
		rules: rules,
	}
}

func (rl ruleList) Match(r *http.Request) (matched bool, err error) {
	return NoCondition(r)
}

func (rl ruleList) Apply(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, rule := range rl.rules {
			matched, err := rule.Match(r)
			switch {
			case matched:
				next = rule.Apply(next)
			case err != nil:
				log.FromContext(r.Context()).ErrorWithFields(log.Fields{"err": err}, "Routing: checking the rule condition for the request")
			}
		}
		next.ServeHTTP(w, r)
	})
}
