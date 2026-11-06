package middleware

import (
	"context"
	"net"
	"net/http"
	"net/netip"
	"strings"
)

type ctxKeyRealIP struct{}

var (
	trueClientIP  = http.CanonicalHeaderKey("True-Client-IP")
	xRealIP       = http.CanonicalHeaderKey("X-Real-IP")
	xForwardedFor = http.CanonicalHeaderKey("X-Forwarded-For")
)

// GetRealIP returns the client IP stored by RealIP.
func GetRealIP(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	ip, _ := ctx.Value(ctxKeyRealIP{}).(string)

	return ip
}

// RealIP stores the client IP for GetRealIP. Header selection matches chi's
// deprecated RealIP: True-Client-IP, then X-Real-IP, then the leftmost
// X-Forwarded-For entry. The first non-empty value is kept only when it
// parses; otherwise the TCP peer is stored. RemoteAddr is left unchanged.
func RealIP(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := realIP(r)
		if ip == "" {
			ip = peerIP(r.RemoteAddr)
		}
		r = r.WithContext(context.WithValue(r.Context(), ctxKeyRealIP{}, ip))
		h.ServeHTTP(w, r)
	})
}

func realIP(r *http.Request) string {
	var ip string

	if tcip := r.Header.Get(trueClientIP); tcip != "" {
		ip = tcip
	} else if xrip := r.Header.Get(xRealIP); xrip != "" {
		ip = xrip
	} else if xff := r.Header.Get(xForwardedFor); xff != "" {
		ip, _, _ = strings.Cut(xff, ",")
	}
	if _, err := netip.ParseAddr(ip); err != nil {
		return ""
	}

	return ip
}

func peerIP(remoteAddr string) string {
	addr, err := netip.ParseAddr(TrimPort(remoteAddr))
	if err != nil {
		return TrimPort(remoteAddr)
	}

	return addr.Unmap().String()
}

func TrimPort(s string) string {
	host, _, err := net.SplitHostPort(s)
	if err != nil {
		return s
	}

	return host
}
