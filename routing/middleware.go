package routing

import "net/http"

func Routing(rules ...Rule) func(next http.Handler) http.Handler {
	return JoinRules(rules...).Apply
}
