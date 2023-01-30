package log

import (
	"bytes"
	"context"
	"net/http"
	"time"

	"github.com/bldsoft/gost/utils"
	"github.com/go-chi/chi/v5/middleware"
)

var RequestInfoCtxKey = &utils.ContextKey{Name: "RequestInfo"}

// ChannelFormatter send requests info to channel.
// To customize request info put RequestInfo in your structure and use it as T.
// Channel formatter puts request info to the context and you can set custom fields after.
// The fields of RequestInfo are filled by ChannelFormatter.
type ChannelFormatter[T any, P RequestInfoPtr[T]] struct {
	requestC          chan<- P
	instanceName      string
	requestInfoCtxKey interface{}
}

func NewChannelFormatter[T any, P RequestInfoPtr[T]](ch chan<- P, instanceName string) *ChannelFormatter[T, P] {
	return &ChannelFormatter[T, P]{requestC: ch, instanceName: instanceName, requestInfoCtxKey: RequestInfoCtxKey}
}

func ChanRequestLogger[T any, P RequestInfoPtr[T]](ch chan<- P, instanseName string) func(next http.Handler) http.Handler {
	return NewRequestLogger(NewChannelFormatter(ch, instanseName))
}

func (f *ChannelFormatter[T, P]) SetRequestInfoContextKey(key interface{}) {
	f.requestInfoCtxKey = key
}

func (f *ChannelFormatter[T, P]) GetRequestInfo(ctx context.Context) P {
	requestInfo, _ := ctx.Value(f.requestInfoCtxKey).(P)
	return requestInfo
}

func (f *ChannelFormatter[T, P]) NewLogEntry(r *http.Request) (middleware.LogEntry, *http.Request) {
	var requestInfo T
	requestInfoPtr := (P)(&requestInfo)
	baseRequestInfo := requestInfoPtr.BaseRequestInfo()

	reqID := middleware.GetReqID(r.Context())
	baseRequestInfo.RequestTime = time.Now().Unix()
	baseRequestInfo.Instance = f.instanceName
	baseRequestInfo.RequestMethod = GetRequestMethodType(r.Method)
	baseRequestInfo.Path = r.URL.Path
	baseRequestInfo.Query = r.URL.RawQuery
	baseRequestInfo.ClientIp = r.RemoteAddr
	baseRequestInfo.UserAgent = r.UserAgent()
	baseRequestInfo.RequestId = reqID
	if baseRequestInfo.RequestMethod == POST && r.ContentLength > 0 {
		baseRequestInfo.Size = uint32(r.ContentLength)
	}

	ctx := context.WithValue(r.Context(), f.requestInfoCtxKey, requestInfoPtr)
	r = r.WithContext(ctx)
	return &ContextChanLoggerEntry[T, P]{requestCh: f.requestC, errBuf: LogRequestErrBufferFromContext(r.Context()), requestInfo: requestInfoPtr}, r
}

type ContextChanLoggerEntry[T any, P RequestInfoPtr[T]] struct {
	requestInfo P
	errBuf      *bytes.Buffer
	requestCh   chan<- P
}

func (l *ContextChanLoggerEntry[T, P]) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	duration := elapsed.Milliseconds()

	if l.requestInfo != nil {
		baseRequestInfo := l.requestInfo.BaseRequestInfo()
		baseRequestInfo.ResponseCode = ResponseCodeType(status)
		if baseRequestInfo.RequestMethod != POST {
			baseRequestInfo.Size = uint32(bytes)
		}
		baseRequestInfo.HandleTime = uint32(duration)

		if l.errBuf != nil {
			baseRequestInfo.Error = l.errBuf.String()
		}
		l.writeInfoToChannel()
	}
}

func (l *ContextChanLoggerEntry[T, P]) writeInfoToChannel() {
	if l.requestInfo != nil {
		select {
		case l.requestCh <- l.requestInfo:
		default:
			go Logger.ErrorWithFields(Fields{"requestId": l.requestInfo}, "Failed to write request info: channel is full.")
		}
	}
}

func (l *ContextChanLoggerEntry[T, P]) Panic(v interface{}, stack []byte) {}
