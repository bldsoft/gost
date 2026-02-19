package log

import (
	"bytes"
	"context"
	"net/http"
	"time"

	"github.com/bldsoft/gost/utils"
	"github.com/bldsoft/gost/utils/exporter"
	"github.com/go-chi/chi/v5/middleware"
)

// To customize request info put RequestInfo in your structure and use it as T.
// Channel formatter puts request info to the context and you can set custom fields after.
// The fields of RequestInfo are filled by ChannelFormatter.
type ExportFormatter[T any, P RequestInfoPtr[T]] struct {
	requestExporter   exporter.Exporter[P]
	instanceName      string
	requestInfoCtxKey interface{}
}

func NewExportFormatter[T any, P RequestInfoPtr[T]](requestExporter exporter.Exporter[P], instanceName string) *ExportFormatter[T, P] {
	return &ExportFormatter[T, P]{requestExporter: requestExporter, instanceName: instanceName, requestInfoCtxKey: RequestInfoCtxKey}
}

func ExportRequestLogger[T any, P RequestInfoPtr[T]](requestExporter exporter.Exporter[P], instanseName string) func(next http.Handler) http.Handler {
	return NewRequestLogger(NewExportFormatter(requestExporter, instanseName))
}

func (f *ExportFormatter[T, P]) SetRequestInfoContextKey(key interface{}) {
	f.requestInfoCtxKey = key
}

func (f *ExportFormatter[T, P]) GetRequestInfo(ctx context.Context) P {
	requestInfo, _ := ctx.Value(f.requestInfoCtxKey).(P)
	return requestInfo
}

func (f *ExportFormatter[T, P]) NewLogEntry(r *http.Request) (middleware.LogEntry, *http.Request) {
	receivedAt := time.Now()

	var requestInfo T
	requestInfoPtr := (P)(&requestInfo)
	baseRequestInfo := requestInfoPtr.BaseRequestInfo()

	reqID := middleware.GetReqID(r.Context())
	baseRequestInfo.RequestTime = receivedAt.Unix()
	baseRequestInfo.RequestTimeMs = uint16(receivedAt.UnixMilli() % 1000)
	baseRequestInfo.Instance = f.instanceName
	baseRequestInfo.RequestMethod = GetRequestMethodType(r.Method)
	if baseRequestInfo.RequestMethod == ERROR {
		Logger.ErrorWithFields(Fields{"method": r.Method, "url": r.URL.Path, "requestID": reqID}, "Request info: bad request method")
	}
	baseRequestInfo.Path = r.URL.Path
	baseRequestInfo.Query = r.URL.RawQuery
	baseRequestInfo.ClientIp = r.RemoteAddr
	// ua := r.UserAgent()
	// baseRequestInfo.UserAgent = ua
	baseRequestInfo.RequestId = reqID
	if utils.IsIn(baseRequestInfo.RequestMethod, POST, PUT) && r.ContentLength > 0 {
		baseRequestInfo.Size = uint32(r.ContentLength)
	}

	ctx := context.WithValue(r.Context(), f.requestInfoCtxKey, requestInfoPtr)
	r = r.WithContext(ctx)
	return &ContextExportFormatterLoggerEntry[T, P]{
		requestExporter: f.requestExporter,
		errBuf:          LogRequestErrBufferFromContext(r.Context()),
		requestInfo:     requestInfoPtr,
		req:             r,
	}, r
}

type ContextExportFormatterLoggerEntry[T any, P RequestInfoPtr[T]] struct {
	requestInfo     P
	errBuf          *bytes.Buffer
	requestExporter exporter.Exporter[P]
	req             *http.Request
}

func (l *ContextExportFormatterLoggerEntry[T, P]) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	duration := elapsed.Microseconds()

	if l.requestInfo != nil {
		baseRequestInfo := l.requestInfo.BaseRequestInfo()
		baseRequestInfo.ResponseCode = ResponseCodeType(status)
		if !utils.IsIn(baseRequestInfo.RequestMethod, POST, PUT) {
			baseRequestInfo.Size = uint32(bytes)
		}
		baseRequestInfo.HandleTime = uint32(duration)

		if l.errBuf != nil {
			baseRequestInfo.Error = l.errBuf.String()
		}

		baseRequestInfo.UserAgent = l.req.UserAgent()

		l.requestExporter.Export(l.requestInfo)
	}
}

func (l *ContextExportFormatterLoggerEntry[T, P]) Panic(v interface{}, stack []byte) {}
