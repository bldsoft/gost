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

// WithLogEntry sets the in-context ServiceLogger for a request.
func WithLogger(r *http.Request, logger *ServiceLogger) *http.Request {
	r = r.WithContext(context.WithValue(r.Context(), LoggerCtxKey, logger))
	return r
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
	return chi.Chain(logger, RequestLogger(f)).Handler
}

type LogFormatter interface {
	// returns Request to put some values into context
	NewLogEntry(r *http.Request) (middleware.LogEntry, *http.Request)
}

func RequestLogger(f LogFormatter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
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
