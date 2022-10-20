package log

type RequestMethodType byte
type ResponseCodeType uint16

const (
	GET RequestMethodType = iota
	POST
	PUT
	DELETE
	OPTIONS
	HEAD

	ERROR
)

type RequestInfoPtr[T any] interface {
	IRequestInfo
	*T
}
type IRequestInfo interface {
	BaseRequestInfo() *RequestInfo
}

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
	default:
		return ERROR
	}
}
