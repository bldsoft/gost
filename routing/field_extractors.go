package routing

import (
	"net/http"
)

type hostExtractor struct{}

func (e hostExtractor) ExtractValue(r *http.Request) string {
	return r.Host
}

var Host = &hostExtractor{}
