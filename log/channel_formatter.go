package log

import (
	"bytes"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type ChannelFormatter struct {
	requestC     chan<- *RequestInfo
	instanseName string
}

func NewChannelFormatter(ch chan<- *RequestInfo, instanseName string) *ChannelFormatter {
	return &ChannelFormatter{requestC: ch, instanseName: instanseName}
}

func ChanRequestLogger(ch chan<- *RequestInfo, instanseName string) func(next http.Handler) http.Handler {
	return NewRequestLogger(NewChannelFormatter(ch, instanseName))
}

func (f *ChannelFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {

	reqID := middleware.GetReqID(r.Context())
	requestInfo := NewRequestInfo()
	requestInfo.Instance = f.instanseName
	requestInfo.RequestMethod = GetRequestMethodType(r.Method)
	requestInfo.Path = r.RequestURI
	requestInfo.ClientIp = r.RemoteAddr
	requestInfo.UserAgent = r.UserAgent()
	requestInfo.RequestId = reqID
	if requestInfo.RequestMethod == POST && r.ContentLength > 0 {
		requestInfo.Size = uint32(r.ContentLength)
	}

	return &ContextChanLoggerEntry{requestCh: f.requestC, errBuf: LogRequestErrBufferFromContext(r.Context()), requestInfo: requestInfo}
}

type ContextChanLoggerEntry struct {
	requestInfo *RequestInfo
	errBuf      *bytes.Buffer
	requestCh   chan<- *RequestInfo
}

func (l *ContextChanLoggerEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	duration := elapsed.Milliseconds()

	if l.requestInfo != nil {
		l.requestInfo.ResponseCode = ResponseCodeType(status)
		if l.requestInfo.RequestMethod != POST {
			l.requestInfo.Size = uint32(bytes)
		}
		l.requestInfo.HandleTime = uint32(duration)

		if l.errBuf != nil {
			l.requestInfo.Error = l.errBuf.String()
		}

		l.writeInfoToChannel()
	}
}

func (l *ContextChanLoggerEntry) writeInfoToChannel() {
	if l.requestInfo != nil {
		select {
		case l.requestCh <- l.requestInfo:
		default:
			go Logger.ErrorWithFields(Fields{"requestId": l.requestInfo.RequestId}, "Failed to write request info: channel is full.")
		}
	}
}

func (l *ContextChanLoggerEntry) Panic(v interface{}, stack []byte) {}
