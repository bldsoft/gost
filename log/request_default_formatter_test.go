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
	router.Use(NewRequestLogger(&DefaultFormatter{HiddenHeaders: []string{"API-Key"}}))
	router.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		handlerKey = r.Header.Get("API-Key")
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("API-Key", "secret")
	req.Header.Set("User-Agent", "agent")
	router.ServeHTTP(httptest.NewRecorder(), req)

	assert.Equal(t, "secret", handlerKey)
	assert.Contains(t, buf.String(), `"Api-Key":["****"]`)
	assert.Contains(t, buf.String(), `"User-Agent":["agent"]`)
	assert.NotContains(t, buf.String(), "secret")
}
