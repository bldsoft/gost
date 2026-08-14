package acl

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetIP(t *testing.T) {
	var m Acl
	tests := []struct {
		name    string
		remote  string
		want    string
		wantErr bool
	}{
		{name: "ipv4 host port", remote: "192.168.1.10:1234", want: "192.168.1.10"},
		{name: "ipv4 bare", remote: "10.0.0.1", want: "10.0.0.1"},
		{name: "ipv6 host port", remote: "[::1]:80", want: "::1"},
		{name: "ipv6 bare", remote: "2001:db8::1", want: "2001:db8::1"},
		{name: "mapped ipv4", remote: "[::ffff:192.168.0.1]:443", want: "192.168.0.1"},
		{name: "invalid", remote: "not-an-ip", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remote
			got, err := m.getIP(req)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, netip.MustParseAddr(tt.want), got)
		})
	}
}

func TestMiddlewareAllowOnly(t *testing.T) {
	h := Acl{Allow: MustIpRangeFromStrings("10.0.0.0/24")}.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	allowed := httptest.NewRequest(http.MethodGet, "/", nil)
	allowed.RemoteAddr = "10.0.0.5:9"
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, allowed)
	assert.Equal(t, http.StatusOK, rr.Code)

	denied := httptest.NewRequest(http.MethodGet, "/", nil)
	denied.RemoteAddr = "11.0.0.1:9"
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, denied)
	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestMiddlewareDenyOnly(t *testing.T) {
	h := Acl{Deny: MustIpRangeFromStrings("192.168.0.1")}.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	denied := httptest.NewRequest(http.MethodGet, "/", nil)
	denied.RemoteAddr = "192.168.0.1:9"
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, denied)
	assert.Equal(t, http.StatusForbidden, rr.Code)

	allowed := httptest.NewRequest(http.MethodGet, "/", nil)
	allowed.RemoteAddr = "10.0.0.1:9"
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, allowed)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestMiddlewareAllowOnlyZeroDeny(t *testing.T) {
	// Deny is the zero IpRange; must not panic.
	h := Acl{Allow: MustIpRangeFromStrings("127.0.0.1")}.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:9"
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestMiddlewareDisabled(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	h := Acl{}.Middleware(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "not-even-an-ip"
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rr.Code)
}
