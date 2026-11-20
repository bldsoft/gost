package log

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type DefaultFormatter struct {
	LogResponseHeaders bool
	HiddenHeaders      []string
	HiddenPathParams   []string
	HiddenQueryParams  []string
}

func NewDefaultFormatter() *DefaultFormatter {
	return &DefaultFormatter{LogResponseHeaders: true}
}

func DefaultRequestLogger() func(next http.Handler) http.Handler {
	return NewRequestLogger(NewDefaultFormatter())
}

func (l *DefaultFormatter) NewLogEntry(r *http.Request) (middleware.LogEntry, *http.Request) {

	entry := &ContextLoggerEntry{
		Logger:             FromContext(r.Context()),
		LogResponseHeaders: l.LogResponseHeaders,
		HiddenHeaders:      l.HiddenHeaders,
	}

	logFields := Fields{}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	path, query, hasQuery := strings.Cut(r.URL.RequestURI(), "?")
	if rctx := chi.RouteContext(r.Context()); rctx != nil && len(l.HiddenPathParams) > 0 {
		tctx := chi.NewRouteContext()
		rctx.Routes.Find(tctx, r.Method, path)
		segments := strings.Split(path, "/")
		for i, key := range tctx.URLParams.Keys {
			if !slices.Contains(l.HiddenPathParams, key) {
				continue
			}
			for j, s := range segments {
				if s == tctx.URLParams.Values[i] {
					segments[j] = "****"
				}
			}
		}
		path = strings.Join(segments, "/")
	}
	if len(l.HiddenQueryParams) > 0 {
		params := strings.Split(query, "&")
		for i, p := range params {
			if k, _, ok := strings.Cut(p, "="); ok && slices.Contains(l.HiddenQueryParams, k) {
				params[i] = k + "=****"
			}
		}
		query = strings.Join(params, "&")
	}
	if hasQuery {
		path += "?" + query
	}
	logFields["uri"] = fmt.Sprintf("%s://%s%s", scheme, r.Host, path)
	logFields["pr"] = r.Proto
	logFields["mt"] = r.Method
	logFields["from"] = r.RemoteAddr
	logFields["hdr"] = maskHeaders(r.Header, l.HiddenHeaders)
	//TODO: trace body
	/*if log.Trace().Enabled() {
		logFields["body"] =
	}*/
	entry.Logger.InfofWithFields(logFields, "%s REQUEST", r.Method)

	return entry, r
}

type ContextLoggerEntry struct {
	//Logger *zerolog.Event
	Logger             *ServiceLogger
	LogResponseHeaders bool
	HiddenHeaders      []string
}

func maskHeaders(header http.Header, hidden []string) http.Header {
	if len(hidden) == 0 {
		return header
	}
	header = header.Clone()
	for _, h := range hidden {
		if header.Get(h) != "" {
			header.Set(h, "****")
		}
	}

	return header
}

func (l *ContextLoggerEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra any) {
	duration := elapsed.Microseconds()

	fields := Fields{
		"resp_status":  status,
		"resp_bytes":   bytes,
		"resp_time_ms": float64(duration) / 1000.0,
	}

	if l.LogResponseHeaders {
		fields["hdr"] = maskHeaders(header, l.HiddenHeaders)
	}

	switch {
	case 200 <= status && status < 400:
		l.Logger.InfoWithFields(fields, "RESPONSE")
	case status >= 500:
		l.Logger.ErrorWithFields(fields, "RESPONSE")
	default:
		l.Logger.WarnWithFields(fields, "RESPONSE")
	}
}

func (l *ContextLoggerEntry) Panic(v any, stack []byte) {
	l.Logger = l.Logger.WithFields(Fields{
		"panic": fmt.Sprintf("%+v", v),
	})
	l.Logger.Error(string(stack))
}
