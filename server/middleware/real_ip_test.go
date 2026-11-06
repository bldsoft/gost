package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRealIP(t *testing.T) {
	tests := []struct {
		name string
		peer string
		set  map[string][]string
		want string
	}{
		{
			name: "true client ip wins",
			peer: "10.0.0.1:9",
			set: map[string][]string{
				"True-Client-IP":  {"198.51.100.1"},
				"X-Real-IP":       {"203.0.113.9"},
				"X-Forwarded-For": {"192.0.2.1"},
			},
			want: "198.51.100.1",
		},
		{
			name: "x real ip is next",
			peer: "10.0.0.1:9",
			set: map[string][]string{
				"X-Real-IP":       {"203.0.113.9"},
				"X-Forwarded-For": {"192.0.2.1"},
			},
			want: "203.0.113.9",
		},
		{
			name: "single xff value is used",
			peer: "10.0.0.1:9",
			set:  map[string][]string{"X-Forwarded-For": {"192.0.2.1"}},
			want: "192.0.2.1",
		},
		{
			name: "leftmost xff entry is used",
			peer: "10.0.0.1:9",
			set:  map[string][]string{"X-Forwarded-For": {"1.2.3.4, 203.0.113.9"}},
			want: "1.2.3.4",
		},
		{
			name: "no header keeps the peer",
			peer: "203.0.113.8:1234",
			want: "203.0.113.8",
		},
		{
			name: "first header value wins",
			peer: "127.0.0.1:9",
			set:  map[string][]string{"X-Real-IP": {"198.51.100.1", "203.0.113.1"}},
			want: "198.51.100.1",
		},
		{
			name: "invalid true client ip does not fall through",
			peer: "10.0.0.1:9",
			set: map[string][]string{
				"True-Client-IP": {"not-an-ip"},
				"X-Real-IP":      {"203.0.113.9"},
			},
			want: "10.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			var remote string
			handler := RealIP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = GetRealIP(r.Context())
				remote = r.RemoteAddr
			}))
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.RemoteAddr = tt.peer
			for header, values := range tt.set {
				for _, value := range values {
					r.Header.Add(header, value)
				}
			}
			handler.ServeHTTP(httptest.NewRecorder(), r)
			require.Equal(t, tt.want, got)
			require.Equal(t, tt.peer, remote)
		})
	}
}
