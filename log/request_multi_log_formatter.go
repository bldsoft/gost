package log

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type MultiLogFormatter struct {
	formatters []LogFormatter
}

func newMultiLogFormatter(formatters ...LogFormatter) *MultiLogFormatter {
	return &MultiLogFormatter{formatters}
}

func MultiLogger(formatters ...LogFormatter) func(next http.Handler) http.Handler {
	return NewRequestLogger(newMultiLogFormatter(formatters...))
}

func (f *MultiLogFormatter) NewLogEntry(r *http.Request) (middleware.LogEntry, *http.Request) {
	entries := make([]middleware.LogEntry, 0, len(f.formatters))
	for _, formatter := range f.formatters {
		var entry LogEntry
		entry, r = formatter.NewLogEntry(r)
		entries = append(entries, entry)
	}

	return &MultiContextLogEntry{entries}, r
}

type MultiContextLogEntry struct {
	contextLoggerEntries []middleware.LogEntry
}

func (e *MultiContextLogEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	for _, entry := range e.contextLoggerEntries {
		entry.Write(status, bytes, header, elapsed, extra)
	}
}

func (e *MultiContextLogEntry) Panic(v interface{}, stack []byte) {
	for _, entry := range e.contextLoggerEntries {
		entry.Panic(v, stack)
	}
}
