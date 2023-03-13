package routing

import (
	"net/http"
)

type ActionModifyHeader struct {
	IncomingRequest bool
	Add             bool
	HeaderName      string
	Value           string
}

func (a ActionModifyHeader) modifyHeader(header http.Header) {
	if a.Add {
		header.Set(a.HeaderName, a.Value)
	} else {
		header.Del(a.HeaderName)
	}
}

func (a ActionModifyHeader) DoBeforeHandle(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request, error) {
	if a.IncomingRequest {
		a.modifyHeader(r.Header)
	}
	return w, r, nil
}

func (a ActionModifyHeader) DoAfterHandle(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request, error) {
	if !a.IncomingRequest {
		a.modifyHeader(w.Header())
	}
	return w, r, nil
}

// for graphql
func (ActionModifyHeader) IsAction() {}
