package routing

import (
	"net/http"
	"path/filepath"
)

type ActionRedirect struct {
	Code        int
	Scheme      string
	Host        string
	ReplacePath string
	PathPrefix  string
	ClearQuery  bool
}

func setIfNotZero[T comparable](dst *T, val T) {
	var zero T
	if val != zero {
		*dst = val
	}
}

func (ar ActionRedirect) Apply(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := *r.URL

		setIfNotZero(&url.Scheme, ar.Scheme)
		setIfNotZero(&url.Host, ar.Host)
		setIfNotZero(&url.Path, ar.ReplacePath)
		if ar.PathPrefix != "" {
			url.Path = filepath.Join(ar.PathPrefix, url.Path)
		}
		if ar.ClearQuery {
			url.RawQuery = ""
		}
		http.Redirect(w, r, url.String(), ar.Code)
	})
}
