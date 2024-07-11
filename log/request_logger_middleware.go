package log

import (
	"context"
	"net/http"
	"time"

	"github.com/bldsoft/gost/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var LoggerCtxKey = &utils.ContextKey{Name: "Logger"}

const ReqIdFieldName = "req_id"

type LogEntry = middleware.LogEntry

func GetLogEntry(ctx context.Context) LogEntry {
	entry, _ := ctx.Value(middleware.LogEntryCtxKey).(LogEntry)
	return entry
}

func GetLogEntryFromRequest(r *http.Request) LogEntry {
	return GetLogEntry(r.Context())
}

// WithLogEntry sets the in-context ServiceLogger for a request.
func WithLogger(r *http.Request, logger *ServiceLogger) *http.Request {
	r = r.WithContext(context.WithValue(r.Context(), LoggerCtxKey, logger))
	return r
}

func WithLoggerCtx(ctx context.Context, logger *ServiceLogger) context.Context {
	return context.WithValue(ctx, LoggerCtxKey, logger)
}

// FromContext extracts Logger from context if exists or return global Logger
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

		logFields := Fields{ReqIdFieldName: reqID}

		logger := Logger.WithFields(logFields)
		next.ServeHTTP(w, WithLogger(r, logger))
	}
	return http.HandlerFunc(fn)
}

func NewRequestLogger(f LogFormatter) func(next http.Handler) http.Handler {
	return chi.Chain(logger, WithLogRequestErrBuffer, requestLogger(f)).Handler
}

// Changed LogFromatter interface from chi.
// It returns request to make possible put values to the context.
type LogFormatter interface {
	NewLogEntry(r *http.Request) (middleware.LogEntry, *http.Request)
}

var NotLoggedEndpoints []string

// requestLogger is a copy of the chi middleware.RequestLogger with changed LogFormatter interface
func requestLogger(f LogFormatter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if utils.IsIn(r.URL.Path, NotLoggedEndpoints...) {
				next.ServeHTTP(w, r)
				return
			}
			entry, r := f.NewLogEntry(r)
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				entry.Write(ww.Status(), ww.BytesWritten(), ww.Header(), time.Since(t1), nil)
			}()

			next.ServeHTTP(ww, middleware.WithLogEntry(r, entry))
		}
		return http.HandlerFunc(fn)
	}
}
