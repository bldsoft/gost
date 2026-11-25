package log

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestDefaultFormatterHiddenHeaders(t *testing.T) {
	var buf bytes.Buffer
	defer func(l ServiceLogger) { Logger = l }(Logger)
	Logger = ServiceLogger{logger: zerolog.New(&buf)}

	var handlerKey string
	router := chi.NewRouter()
	router.Use(NewRequestLogger(&DefaultFormatter{LogResponseHeaders: true, HiddenHeaders: []string{"API-Key", "Set-Cookie"}}))
	router.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		handlerKey = r.Header.Get("API-Key")
		w.Header().Set("Set-Cookie", "session=secret")
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("API-Key", "secret")
	req.Header.Set("User-Agent", "agent")
	router.ServeHTTP(httptest.NewRecorder(), req)

	assert.Equal(t, "secret", handlerKey)
	assert.Contains(t, buf.String(), `"Api-Key":["****"]`)
	assert.Contains(t, buf.String(), `"User-Agent":["agent"]`)
	assert.Contains(t, buf.String(), `"Set-Cookie":["****"]`)
	assert.NotContains(t, buf.String(), "secret")
}

func TestDefaultFormatterHiddenPathParams(t *testing.T) {
	var buf bytes.Buffer
	defer func(l ServiceLogger) { Logger = l }(Logger)
	Logger = ServiceLogger{logger: zerolog.New(&buf)}

	var handlerToken string
	router := chi.NewRouter()
	router.Use(NewRequestLogger(&DefaultFormatter{HiddenPathParams: []string{"token"}}))
	router.Route("/media", func(r chi.Router) {
		r.Get("/{token}/{file}", func(w http.ResponseWriter, r *http.Request) {
			handlerToken = chi.URLParam(r, "token")
		})
	})

	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/media/secret/file.mpd", nil))

	assert.Equal(t, "secret", handlerToken)
	assert.Contains(t, buf.String(), `"uri":"http://example.com/media/****/file.mpd"`)
	assert.NotContains(t, buf.String(), "secret")
}

func TestDefaultFormatterHiddenQueryParams(t *testing.T) {
	var buf bytes.Buffer
	defer func(l ServiceLogger) { Logger = l }(Logger)
	Logger = ServiceLogger{logger: zerolog.New(&buf)}

	var handlerToken string
	router := chi.NewRouter()
	router.Use(NewRequestLogger(&DefaultFormatter{HiddenQueryParams: []string{"token"}}))
	router.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		handlerToken = r.URL.Query().Get("token")
	})

	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/file.mpd?token=secret&b=1", nil))

	assert.Equal(t, "secret", handlerToken)
	assert.Contains(t, buf.String(), `"uri":"http://example.com/file.mpd?token=****&b=1"`)
	assert.NotContains(t, buf.String(), "secret")
}
