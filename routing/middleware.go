package routing

import "net/http"

func Routing(rules ...IRule) func(next http.Handler) http.Handler {
	return JoinRules(rules...).Apply
}
