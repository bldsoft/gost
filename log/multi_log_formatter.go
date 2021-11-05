package log

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type MultiLogFormatter struct {
	formatters []middleware.LogFormatter
}

func newMultiLogFormatter(formatters ...middleware.LogFormatter) *MultiLogFormatter {
	return &MultiLogFormatter{formatters}
}

func MultiLogger(formatters ...middleware.LogFormatter) func(next http.Handler) http.Handler {
	return NewRequestLogger(newMultiLogFormatter(formatters...))
}

func (f *MultiLogFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	entries := make([]middleware.LogEntry, 0, len(f.formatters))
	for _, formatter := range f.formatters {
		entries = append(entries, formatter.NewLogEntry(r))
	}

	return &MultiContextLogEntry{entries}
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