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

func (ar ActionRedirect) redirect(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {
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
	return w, r
}

func (ar ActionRedirect) DoBeforeHandle(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request, error) {
	if ar.IncomingRequest {
		w, r := ar.redirect(w, r)
		return w, r, ErrStopHandling
	}
	return w, r, nil
}

func (ar ActionRedirect) DoAfterHandle(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request, error) {
	if !ar.IncomingRequest {
		w, r := ar.redirect(w, r)
		return w, r, nil
	}
	return w, r, nil
}

// for graphql
func (ActionRedirect) IsAction() {}
