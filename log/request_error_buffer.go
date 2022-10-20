package log

import (
	"bufio"
	"bytes"
	"context"
	"net"
	"net/http"

	"github.com/bldsoft/gost/utils"
	"github.com/go-chi/chi/v5/middleware"
)

var requestErrorBufferCtxKey = &utils.ContextKey{Name: "RequestErrorBuf"}

type WrapResponseWriterLogErr interface {
	middleware.WrapResponseWriter
	WriteRequestInfoErr(s string)
	ErrBuffer() *bytes.Buffer
	Error() string
}

type wrapResponseWriterLogErr struct {
	middleware.WrapResponseWriter
	errBuf bytes.Buffer
}

type hijackWriterLogErr struct {
	*wrapResponseWriterLogErr
}

func (w *hijackWriterLogErr) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj := w.wrapResponseWriterLogErr.WrapResponseWriter.(http.Hijacker)
	return hj.Hijack()
}

func NewWrapResponseWriterLogErr(w http.ResponseWriter, protoMajor int) WrapResponseWriterLogErr {
	wrapWriter := middleware.NewWrapResponseWriter(w, protoMajor)
	wrapWriterLogErr := &wrapResponseWriterLogErr{WrapResponseWriter: wrapWriter, errBuf: bytes.Buffer{}}
	if _, ok := wrapWriter.(http.Hijacker); ok {
		return &hijackWriterLogErr{wrapResponseWriterLogErr: wrapWriterLogErr}
	} else {
		return wrapWriterLogErr
	}
}

func (w *wrapResponseWriterLogErr) WriteRequestInfoErr(s string) {
	w.errBuf.WriteString(s)
}

func (w *wrapResponseWriterLogErr) ErrBuffer() *bytes.Buffer {
	return &w.errBuf
}

func (w *wrapResponseWriterLogErr) Error() string {
	return w.errBuf.String()
}

// WrapResponseWriterLogErr returns writer that can be used to write to a RequestInfo.Error
func AsResponseWriterLogErr(w http.ResponseWriter) (WrapResponseWriterLogErr, bool) {
	for {
		if wrapW, ok := w.(middleware.WrapResponseWriter); ok {
			if result, ok := wrapW.(WrapResponseWriterLogErr); ok {
				return result, ok
			} else {
				w = wrapW.Unwrap()
			}
		} else {
			return nil, false
		}
	}
}

// WithLogRequestErrBuffer put a buffer into the request context.
// The buffer will then be used by a log middleware with ChannelFormatter to write a RequestInfo.Error
func LogRequestErrBufferFromContext(ctx context.Context) *bytes.Buffer {
	if ctx != nil {
		if buf, ok := ctx.Value(requestErrorBufferCtxKey).(*bytes.Buffer); ok {
			return buf
		}
	}
	return nil
}

// WithLogRequestErrBuffer put a buffer into the request context.
// A logging middleware with ChannelFormatter will then use the buffer to write a RequestInfo.Error
func WithLogRequestErrBuffer(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ww := NewWrapResponseWriterLogErr(w, r.ProtoMajor)
		r = r.WithContext(context.WithValue(r.Context(), requestErrorBufferCtxKey, ww.ErrBuffer()))
		next.ServeHTTP(ww, r)
	}
	return http.HandlerFunc(fn)
}
