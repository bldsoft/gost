package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Key to use when setting the real IP.
type ctxKeyRealIP int

const RealIPKey ctxKeyRealIP = 0

func GetRealIP(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if realIP, ok := ctx.Value(RealIPKey).(string); ok {
		return realIP
	}
	return ""
}

func WithRealIP(ctx context.Context, realIP string) context.Context {
	return context.WithValue(ctx, RealIPKey, realIP)
}

// RealIP is a middleware that sets a http.Request's RemoteAddr to the results
// of parsing either the X-Real-IP header or the X-Forwarded-For header (in that
// order). It also put the value into the request context.
func RealIP(h http.Handler) http.Handler {
	return chi.Chain(middleware.RealIP, injectRealIP).Handler(h)
}

func injectRealIP(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(WithRealIP(r.Context(), r.RemoteAddr))
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
