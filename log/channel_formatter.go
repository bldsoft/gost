package log

import (
	"bytes"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type ChannelFormatter[T any, P RequestInfoPtr[T]] struct {
	requestC     chan<- P
	instanceName string
}

func NewChannelFormatter[T any, P RequestInfoPtr[T]](ch chan<- P, instanceName string) *ChannelFormatter[T, P] {
	return &ChannelFormatter[T, P]{requestC: ch, instanceName: instanceName}
}

func ChanRequestLogger[T any, P RequestInfoPtr[T]](ch chan<- P, instanseName string) func(next http.Handler) http.Handler {
	return NewRequestLogger(NewChannelFormatter(ch, instanseName))
}

func (f *ChannelFormatter[T, P]) NewLogEntry(r *http.Request) middleware.LogEntry {
	var requestInfo T
	requestInfoPtr := (P)(&requestInfo)
	baseRequestInfo := requestInfoPtr.BaseRequestInfo()

	reqID := middleware.GetReqID(r.Context())
	baseRequestInfo.RequestTime = time.Now().Unix()
	baseRequestInfo.Instance = f.instanceName
	baseRequestInfo.RequestMethod = GetRequestMethodType(r.Method)
	baseRequestInfo.Path = r.RequestURI
	baseRequestInfo.ClientIp = r.RemoteAddr
	baseRequestInfo.UserAgent = r.UserAgent()
	baseRequestInfo.RequestId = reqID
	if baseRequestInfo.RequestMethod == POST && r.ContentLength > 0 {
		baseRequestInfo.Size = uint32(r.ContentLength)
	}

	return &ContextChanLoggerEntry[T, P]{requestCh: f.requestC, errBuf: LogRequestErrBufferFromContext(r.Context()), requestInfo: requestInfoPtr}
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
