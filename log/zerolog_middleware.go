package log

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	//"github.com/rs/zerolog"
)

type StructuredLogger struct{}

func newLogger() *StructuredLogger {
	return &StructuredLogger{}
}

func NewRequestLogger() func(next http.Handler) http.Handler {
	return middleware.RequestLogger(newLogger())
}

func (l *StructuredLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
	reqID := middleware.GetReqID(r.Context())

	logFields := Fields{"req_id": reqID}

	entry := &ContextLoggerEntry{
		Logger: Logger.WithFields(logFields),
	}
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
	Logger.InfoWithFields(logFields, "REQUEST")

	return entry
}

type ContextLoggerEntry struct {
	//Logger *zerolog.Event
	Logger ServiceLogger
}

func (l *ContextLoggerEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	duration := elapsed.Microseconds()

	fields := Fields{
		"resp_status":  status,
		"resp_bytes":   bytes,
		"resp_time_ms": float64(duration) / 1000.0,
	}

	switch status {
	case http.StatusOK:
		l.Logger.InfoWithFields(fields, "RESPONSE")
	case http.StatusInternalServerError:
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

//GetRequestLogEntry extracts log entry from request's context
func GetRequestLogEntry(r *http.Request) *ContextLoggerEntry {
	entry, _ := middleware.GetLogEntry(r).(*ContextLoggerEntry)
	return entry
}

//FromContext extracts Logger from context if exists or return global Logger
func FromContext(ctx context.Context) *ServiceLogger {
	if ctx != nil {
		if entry, ok := ctx.Value(middleware.LogEntryCtxKey).(*ContextLoggerEntry); ok {
			return &entry.Logger
		}
	}
	return &Logger
}
