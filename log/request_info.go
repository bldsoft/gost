package log

type (
	RequestMethodType byte
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
	RequestMethod RequestMethodType `requestinfo:"RequestMethod"`
	Path          string            `requestinfo:"Path"`
	Query         string            `requestinfo:"Query"`
	ClientIp      string            `requestinfo:"ClientIp"`
	UserAgent     string            `requestinfo:"UserAgent"`
	ResponseCode  ResponseCodeType  `requestinfo:"ResponseCode"`
	Size          uint32            `requestinfo:"Size"`
	RequestTime   int64             `requestinfo:"RequestTime"`
	HandleTime    uint32            `requestinfo:"HandleTime"`
	Instance      string            `requestinfo:"Instance"`
	RequestId     string            `requestinfo:"RequestId"`
	Error         string            `requestinfo:"Error"`
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
