package log

import (
	"net/http"

	"github.com/bldsoft/gost/utils"
	"github.com/bldsoft/gost/utils/exporter"
)

var RequestInfoCtxKey = &utils.ContextKey{Name: "RequestInfo"}

// ChannelFormatter send requests info to channel.
// To customize request info put RequestInfo in your structure and use it as T.
// Channel formatter puts request info to the context and you can set custom fields after.
// The fields of RequestInfo are filled by ChannelFormatter.

func NewChannelFormatter[T any, P RequestInfoPtr[T]](ch chan<- P, instanceName string) *ExportFormatter[T, P] {
	return NewExportFormatter[T, P](exporter.Chan[P](ch), instanceName)
}

func ChanRequestLogger[T any, P RequestInfoPtr[T]](ch chan<- P, instanseName string) func(next http.Handler) http.Handler {
	return NewRequestLogger(NewChannelFormatter(ch, instanseName))
}
