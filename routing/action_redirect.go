package routing

import (
	"net/http"
	"path/filepath"
)

type ActionRedirect struct {
	IncomingRequest bool   `json:"incomingRequest" bson:"incomingRequest"`
	Code            int    `json:"code" bson:"code"`
	Scheme          string `json:"scheme,omitempty" bson:"scheme,omitempty"`
	Host            string `json:"host,omitempty" bson:"host,omitempty"`
	ReplacePath     string `json:"replacePath,omitempty" bson:"replacePath,omitempty"`
	PathPrefix      string `json:"pathPrefix,omitempty" bson:"pathPrefix,omitempty"`
	ClearQuery      bool   `json:"clearQuery,omitempty" bson:"clearQuery,omitempty"`
}

func setIfNotZero[T comparable](dst *T, val T) {
	var zero T
	if val != zero {
		*dst = val
	}
}

func (ar ActionRedirect) Apply(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !ar.IncomingRequest {
			h.ServeHTTP(w, r)
		}
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

// for graphql
func (ActionRedirect) IsAction() {}
