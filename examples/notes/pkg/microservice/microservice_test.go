package microservice_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	gost "github.com/bldsoft/gost/config"
	"github.com/bldsoft/gost/examples/notes/pkg/config"
	"github.com/bldsoft/gost/examples/notes/pkg/microservice"
)

const testAPIKey = "test-key"

// The router is built exactly as server.Server does it, minus the listener:
// Server.Start registers on http.DefaultServeMux, so tests never call it.
func newRouter(t *testing.T) http.Handler {
	t.Helper()
	var cfg config.Config
	require.NoError(t, gost.SetDefaults(&cfg))
	cfg.AdminAPIKey = testAPIKey
	require.NoError(t, gost.Validate(&cfg))

	r := chi.NewRouter()
	r.Group(microservice.New(cfg).BuildRoutes)

	return r
}

func get(h http.Handler, target string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	return rec
}

func TestStandardEndpoints(t *testing.T) {
	h := newRouter(t)

	t.Run("ping", func(t *testing.T) {
		rec := get(h, "/notes/ping", nil)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "pong", rec.Body.String())
	})

	t.Run("version", func(t *testing.T) {
		rec := get(h, "/notes/version", nil)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	})

	t.Run("env requires the api key", func(t *testing.T) {
		rec := get(h, "/notes/env", nil)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("env masks secrets", func(t *testing.T) {
		rec := get(h, "/notes/env", map[string]string{microservice.APIKeyHeader: testAPIKey})
		require.Equal(t, http.StatusOK, rec.Code)
		require.Contains(t, rec.Body.String(), "ServiceName: notes")
		require.Contains(t, rec.Body.String(), `AdminAPIKey: '****'`)
		require.NotContains(t, rec.Body.String(), testAPIKey)
	})

	t.Run("unknown path", func(t *testing.T) {
		rec := get(h, "/notes/nope", nil)
		require.Equal(t, http.StatusNotFound, rec.Code)
	})
}
