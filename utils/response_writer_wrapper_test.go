package utils

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWrapResponseWriterDoubleRedirect(t *testing.T) {
	r := httptest.NewRequest("GET", "http://0", nil)
	redirects := []string{
		"http://1",
		"http://2",
	}

	cases := []struct {
		name             string
		writerWrap       func(w http.ResponseWriter) http.ResponseWriter
		expectedLocation string
	}{
		{
			name:             "raw response writer",
			writerWrap:       func(w http.ResponseWriter) http.ResponseWriter { return w },
			expectedLocation: redirects[0],
		},
		{
			name:             "wrapped response writer",
			writerWrap:       func(w http.ResponseWriter) http.ResponseWriter { return WrapResponseWriter(w) },
			expectedLocation: redirects[len(redirects)-1],
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ww := tc.writerWrap(w)
			for _, url := range redirects {
				http.Redirect(ww, r, url, http.StatusMovedPermanently)
			}
			require.Equal(t, tc.expectedLocation, w.Result().Header["Location"][0])
		})
	}
}

func BenchmarkWrapResponseWriterServer(b *testing.B) {
	cases := []struct {
		name       string
		middleware func(next http.Handler) http.Handler
	}{
		{"raw response writer", func(next http.Handler) http.Handler {
			return next
		}},
		{"wrapped response writer", func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ww := WrapResponseWriter(w)
				defer ww.Flush()
				next.ServeHTTP(ww, r)
			})
		}},
	}
	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			const KB = 1024
			const MB = 1024 * KB
			body := make([]byte, 5*MB)
			var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write(body)
			})
			handler = c.middleware(handler)
			srv := httptest.NewServer(handler)
			defer srv.Close()

			r, _ := http.NewRequest("GET", srv.URL, nil)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				resp, _ := http.DefaultClient.Do(r)
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
		})
	}
}
