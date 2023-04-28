package log

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type DefaultFormatter struct {
	LogResponseHeaders bool
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
	}

	logFields := Fields{}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	logFields["uri"] = fmt.Sprintf("%s://%s%s", scheme, r.Host, r.URL.RequestURI())
	logFields["pr"] = r.Proto
	logFields["mt"] = r.Method
	logFields["from"] = r.RemoteAddr
	logFields["hdr"] = r.Header
	//TODO: trace body
	/*if log.Trace().Enabled() {
		logFields["body"] =
	}*/
	entry.Logger.InfoWithFields(logFields, "REQUEST")

	return entry, r
}

type ContextLoggerEntry struct {
	//Logger *zerolog.Event
	Logger             *ServiceLogger
	LogResponseHeaders bool
}

func (l *ContextLoggerEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	duration := elapsed.Microseconds()

	fields := Fields{
		"resp_status":  status,
		"resp_bytes":   bytes,
		"resp_time_ms": float64(duration) / 1000.0,
	}

	if l.LogResponseHeaders {
		fields["hdr"] = header
	}

	switch {
	case 200 <= status && status < 300:
		l.Logger.InfoWithFields(fields, "RESPONSE")
	case status == http.StatusInternalServerError:
		l.Logger.ErrorWithFields(fields, "RESPONSE")
	default:
		l.Logger.WarnWithFields(fields, "RESPONSE")
	}
}

func (l *ContextLoggerEntry) Panic(v interface{}, stack []byte) {
	l.Logger = l.Logger.WithFields(Fields{
		"panic": fmt.Sprintf("%+v", v),
	})
	l.Logger.Error(string(stack))
}
