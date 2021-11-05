package log

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bldsoft/gost/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	//"github.com/rs/zerolog"
)

var LoggerCtxKey = &utils.ContextKey{Name: "Logger"}

// WithLogEntry sets the in-context ServiceLogger for a request.
func WithLogger(r *http.Request, logger *ServiceLogger) *http.Request {
	r = r.WithContext(context.WithValue(r.Context(), LoggerCtxKey, logger))
	return r
}

//FromContext extracts Logger from context if exists or return global Logger
func FromContext(ctx context.Context) *ServiceLogger {
	if ctx != nil {
		if logger, ok := ctx.Value(LoggerCtxKey).(*ServiceLogger); ok {
			return logger
		}
	}
	return &Logger
}

// logger is a middleware that injects a Logger into the context
func logger(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		reqID := middleware.GetReqID(r.Context())

		logFields := Fields{"req_id": reqID}

		logger := Logger.WithFields(logFields)
		next.ServeHTTP(w, WithLogger(r, logger))
	}
	return http.HandlerFunc(fn)
}

func NewRequestLogger(f middleware.LogFormatter) func(next http.Handler) http.Handler {
	return chi.Chain(logger, WithLogRequestErrBuffer, middleware.RequestLogger(f)).Handler
}

type DefaultFormatter struct{}

func NewDefaultFormatter() *DefaultFormatter {
	return &DefaultFormatter{}
}

func DefaultRequestLogger() func(next http.Handler) http.Handler {
	return NewRequestLogger(NewDefaultFormatter())
}

func (l *DefaultFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {

	entry := &ContextLoggerEntry{
		Logger: FromContext(r.Context()),
	}

	logFields := Fields{}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	logFields["uri"] = fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)
	logFields["pr"] = r.Proto
	logFields["mt"] = r.Method
	logFields["from"] = r.RemoteAddr
	logFields["hdr"] = r.Header
	//TODO: trace body
	/*if log.Trace().Enabled() {
		logFields["body"] =
	}*/
	entry.Logger.InfoWithFields(logFields, "REQUEST")

	return entry
}

type ContextLoggerEntry struct {
	//Logger *zerolog.Event
	Logger *ServiceLogger
}

func (l *ContextLoggerEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	duration := elapsed.Microseconds()

	fields := Fields{
		"resp_status":  status,
		"resp_bytes":   bytes,
		"resp_time_ms": float64(duration) / 1000.0,
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
