package log

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type RequestMethodType byte
type ResponseCodeType uint16

const (
	GET RequestMethodType = iota
	POST
	PUT
	DELETE
	OPTIONS

	ERROR
)

type RequestInfo struct {
	RequestMethod RequestMethodType
	Path          string
	ClientIp      string
	UserAgent     string
	ResponseCode  ResponseCodeType
	Size          uint32
	RequestTime   int64
	HandleTime    uint32
	Instance      string
	RequestId     string
	Error         string
}

func NewRequestInfo() *RequestInfo {
	return &RequestInfo{RequestTime: time.Now().Unix()}
}

func GetRequestMethodType(type_str string) RequestMethodType {
	switch type_str {
	case "GET":
		return GET
	case "POST":
		return POST
	case "PUT":
		return PUT
	case "DELETE":
		return DELETE
	case "OPTIONS":
		return OPTIONS
	default:
		return ERROR
	}
}

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

	return &ContextChanLoggerEntry{requestCh: f.requestC, requestInfo: requestInfo}
}

type ContextChanLoggerEntry struct {
	requestInfo *RequestInfo
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
