package log

type (
	RequestMethodType int8
	ResponseCodeType  uint16
)

const (
	GET RequestMethodType = iota
	POST
	PUT
	DELETE
	OPTIONS
	HEAD
	PATCH
	ORIGIN_GET

	ERROR = -1
)

type RequestInfoPtr[T any] interface {
	IRequestInfo
	*T
}
type IRequestInfo interface {
	BaseRequestInfo() *RequestInfo
}

type RequestInfo struct {
	Path          string
	Query         string
	ClientIp      string
	UserAgent     string
	Instance      string
	RequestId     string
	Error         string
	RequestTime   int64
	Size          uint32
	HandleTime    uint32
	ResponseCode  ResponseCodeType
	RequestMethod RequestMethodType
}

func (r *RequestInfo) BaseRequestInfo() *RequestInfo {
	return r
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
	case "HEAD":
		return HEAD
	case "PATCH":
		return PATCH
	case "ORIGIN_GET":
		return ORIGIN_GET
	default:
		return ERROR
	}
}
