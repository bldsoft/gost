package log

import (
	"bytes"
	"context"
	"net/http"

	"github.com/bldsoft/gost/utils"
	"github.com/go-chi/chi/v5/middleware"
)

var requestErrorBufferCtxKey = &utils.ContextKey{Name: "RequestErrorBuf"}

type WrapResponseWriterLogErr struct {
	middleware.WrapResponseWriter
	ErrBuf bytes.Buffer
}

func NewWrapResponseWriterLogErr(w http.ResponseWriter, protoMajor int) *WrapResponseWriterLogErr {
	return &WrapResponseWriterLogErr{WrapResponseWriter: middleware.NewWrapResponseWriter(w, protoMajor), ErrBuf: bytes.Buffer{}}
}

func (w *WrapResponseWriterLogErr) WriteRequestInfoErr(s string) {
	w.ErrBuf.WriteString(s)
}

func (w *WrapResponseWriterLogErr) Error() string {
	return w.ErrBuf.String()
}

func AsResponseWriterLogErr(w http.ResponseWriter) (*WrapResponseWriterLogErr, bool) {
	for {
		if wrapW, ok := w.(middleware.WrapResponseWriter); ok {
			if result, ok := wrapW.(*WrapResponseWriterLogErr); ok {
				return result, ok
			} else {
				w = wrapW.Unwrap()
			}
		} else {
			return nil, false
		}
	}
}

func LogRequestErrBufferFromContext(ctx context.Context) *bytes.Buffer {
	if ctx != nil {
		if buf, ok := ctx.Value(requestErrorBufferCtxKey).(*bytes.Buffer); ok {
			return buf
		}
	}
	return nil
}

func WithLogRequestErrBuffer(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ww := NewWrapResponseWriterLogErr(w, r.ProtoMajor)
		r = r.WithContext(context.WithValue(r.Context(), requestErrorBufferCtxKey, &ww.ErrBuf))
		next.ServeHTTP(ww, r)
	}
	return http.HandlerFunc(fn)
}
