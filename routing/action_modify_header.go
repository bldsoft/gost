package routing

import (
	"net/http"

	"github.com/bldsoft/gost/utils"
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

func (a ActionModifyHeader) Apply(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.IncomingRequest {
			a.modifyHeader(r.Header)
		} else {
			ww := utils.WrapResponseWriter(w)
			defer ww.Flush()
			w = ww
		}

		h.ServeHTTP(w, r)

		if !a.IncomingRequest {
			a.modifyHeader(w.Header())
		}
	})
}

// for graphql
func (ActionModifyHeader) IsAction() {}
