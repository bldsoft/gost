package auth

import (
	"net/http"

	"github.com/bldsoft/gost/log"
)

func ApiKeyMiddleware(header, apiKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			passedApiKey := r.Header.Get(header)
			if apiKey != passedApiKey {
				log.FromContext(r.Context()).DebugWithFields(log.Fields{header: passedApiKey}, "Empty or invalid API key")
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
