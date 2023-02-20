package routing

import (
	"net/http"
)

type HostExtractor struct{}

func (e HostExtractor) ExtractValue(r *http.Request) string { return r.Host }

var Host = HostExtractor{}
