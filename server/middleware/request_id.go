package middleware

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

var QueryReqIDParam = "req-id"
var RequestIDHeader = middleware.RequestIDHeader

func RequestID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if requestID := r.URL.Query().Get(QueryReqIDParam); requestID != "" {
			r.Header.Set(RequestIDHeader, requestID)
		}
		middleware.RequestID(next).ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
