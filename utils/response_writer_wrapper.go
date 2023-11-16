package utils

import (
	"bufio"
	"bytes"
	"errors"
	"net"
	"net/http"

	"github.com/felixge/httpsnoop"
)

// state struct
type state struct {
	Body bytes.Buffer
	Code int
}

// ResponseWriter wrapper to allows middleware to set headers without worrying about the response header being written already.
type ResponseWriter struct {
	http.ResponseWriter
	*state

	original http.ResponseWriter
}

// Flush the response
func (rw *ResponseWriter) Flush() (int, error) {
	if rw.state.Code > 0 {
		rw.original.WriteHeader(rw.state.Code)
	}
	return rw.original.Write(rw.state.Body.Bytes())
}

func (rw *ResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack not supported")
	}
	return h.Hijack()
}

func WrapResponseWriter(w http.ResponseWriter) *ResponseWriter {
	state := new(state)
	responseWriter := httpsnoop.Wrap(w, httpsnoop.Hooks{
		WriteHeader: func(_ httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
			return func(code int) {
				state.Code = code
			}
		},
		Write: func(_ httpsnoop.WriteFunc) httpsnoop.WriteFunc {
			return func(p []byte) (int, error) {
				return state.Body.Write(p)
			}
		},
	})
	return &ResponseWriter{
		responseWriter,
		state,
		w,
	}
}

// Unwrap the response writer
func UnwrapResponseWriter(w http.ResponseWriter) (rw *ResponseWriter, ok bool) {
	rw, ok = w.(*ResponseWriter)
	return rw, ok
}
