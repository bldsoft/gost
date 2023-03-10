package test

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bldsoft/gost/routing"
	"github.com/stretchr/testify/assert"
)

func TestRouting(t *testing.T) {
	type args struct {
		r       *http.Request
		rule    routing.IRule
		handler func(t *testing.T) http.Handler
	}
	tests := []struct {
		name string
		args args
		want *http.Response
	}{
		{
			name: "incoming redirect",
			args: args{
				r:       httptest.NewRequest(http.MethodGet, "http://example.com", nil),
				rule:    routing.NewRule(nil, routing.ActionRedirect{IncomingRequest: true, Code: http.StatusMovedPermanently, Host: "go.dev"}),
				handler: OkHandler,
			},
			want: &http.Response{
				StatusCode: http.StatusMovedPermanently,
				Header: http.Header{
					"Location": []string{"http://go.dev"},
				},
			},
		},
		{
			name: "outgoing redirect",
			args: args{
				r:       httptest.NewRequest(http.MethodGet, "http://example.com", nil),
				rule:    routing.NewRule(nil, routing.ActionRedirect{IncomingRequest: false, Code: http.StatusMovedPermanently, Host: "go.dev"}),
				handler: OkHandler,
			},
			want: &http.Response{
				StatusCode: http.StatusMovedPermanently,
				Header: http.Header{
					"Location": []string{"http://go.dev"},
				},
			},
		},
		{
			name: "multiple incoming redirects with first incoming",
			args: args{
				r: httptest.NewRequest(http.MethodGet, "http://example.com", nil),
				rule: routing.NewRule(nil,
					routing.JoinActions(
						routing.ActionRedirect{IncomingRequest: true, Code: http.StatusMovedPermanently, Host: "incoming1"},
						routing.ActionRedirect{IncomingRequest: true, Code: http.StatusMovedPermanently, Host: "incoming2"},
						routing.ActionRedirect{IncomingRequest: false, Code: http.StatusMovedPermanently, Host: "outgoing"},
					)),
				handler: OkHandler,
			},
			want: &http.Response{
				StatusCode: http.StatusMovedPermanently,
				Header: http.Header{
					"Location": []string{"http://incoming1"},
				},
			},
		},
		{
			name: "multiple incoming redirects with first outgoing",
			args: args{
				r: httptest.NewRequest(http.MethodGet, "http://example.com", nil),
				rule: routing.NewRule(nil,
					routing.JoinActions(
						routing.ActionRedirect{IncomingRequest: false, Code: http.StatusMovedPermanently, Host: "outgoing1"},
						routing.ActionRedirect{IncomingRequest: false, Code: http.StatusMovedPermanently, Host: "outgoing2"},
						routing.ActionRedirect{IncomingRequest: true, Code: http.StatusMovedPermanently, Host: "incoming"},
					)),
				handler: OkHandler,
			},
			want: &http.Response{
				StatusCode: http.StatusMovedPermanently,
				Header: http.Header{
					"Location": []string{"http://outgoing1"},
				},
			},
		},
		{
			name: "set request header",
			args: args{
				r: httptest.NewRequest(http.MethodGet, "http://example.com", nil),
				rule: routing.NewRule(nil,
					routing.JoinActions(
						routing.ActionModifyHeader{IncomingRequest: true, Add: true, HeaderName: "Header", Value: "value"},
					)),
				handler: func(t *testing.T) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						assert.Equal(t, http.Header{
							"Header": []string{"value"},
						}, r.Header)
						w.WriteHeader(http.StatusOK)
					})
				},
			},
			want: &http.Response{
				StatusCode: http.StatusOK,
			},
		},
		{
			name: "set response header",
			args: args{
				r: httptest.NewRequest(http.MethodGet, "http://example.com", nil),
				rule: routing.NewRule(nil,
					routing.JoinActions(
						routing.ActionModifyHeader{IncomingRequest: false, Add: true, HeaderName: "Header", Value: "value"},
					)),
				handler: OkHandler,
			},
			want: &http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Header": []string{"value"},
				},
			},
		},
		{
			name: "action order",
			args: args{
				r: httptest.NewRequest(http.MethodGet, "http://example.com", nil),
				rule: routing.JoinRules(
					routing.NewRule(nil, WriteBodyAction{incoming: true, str: "1"}),
					routing.NewRule(nil, WriteBodyAction{incoming: true, str: "2"}),
					routing.JoinRules(
						routing.NewRule(nil, WriteBodyAction{incoming: true, str: "3"}),
						routing.NewRule(nil, WriteBodyAction{incoming: true, str: "4"}),
						routing.JoinRules(
							routing.NewRule(nil, WriteBodyAction{incoming: true, str: "5"}),
							routing.NewRule(nil, WriteBodyAction{incoming: false, str: "6"}),
						),
						routing.NewRule(nil, WriteBodyAction{incoming: false, str: "7"}),
						routing.NewRule(nil, WriteBodyAction{incoming: false, str: "8"}),
					),
					routing.JoinRules(
						routing.NewRule(nil, WriteBodyAction{incoming: true, str: "9"}),
						routing.NewRule(nil, WriteBodyAction{incoming: true, str: "_10"}),
						routing.JoinRules(
							routing.NewRule(nil, WriteBodyAction{incoming: true, str: "_11"}),
							routing.NewRule(nil, WriteBodyAction{incoming: false, str: "_12"}),
						),
						routing.NewRule(nil, WriteBodyAction{incoming: false, str: "_13"}),
						routing.NewRule(nil, WriteBodyAction{incoming: false, str: "_14"}),
					),
					routing.NewRule(nil, WriteBodyAction{incoming: false, str: "_15"}),
				),
				handler: OkHandler,
			},
			want: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("123459_10_11OK678_12_13_14_15")),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := routing.Routing(tt.args.rule)(tt.args.handler(t))
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, tt.args.r)
			ResponseEqual(t, tt.want, w.Result())
		})
	}
}

func OkHandler(t *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}

func ResponseEqual(t *testing.T, want, got *http.Response) {
	assert.Equal(t, want.StatusCode, got.StatusCode)

	if want.Body != nil {
		wantBody := ResponseBody(t, want)
		assert.Equal(t, wantBody, ResponseBody(t, got))
	}

	for name, value := range want.Header {
		assert.Equal(t, value, got.Header[name])
	}
}

func ResponseBody(t *testing.T, resp *http.Response) []byte {
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	return data
}

type WriteBodyAction struct {
	incoming bool
	str      string
}

func (a WriteBodyAction) Incoming() bool {
	return a.incoming
}

func (a WriteBodyAction) Apply(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.incoming {
			w.Write([]byte(a.str))
		}
		h.ServeHTTP(w, r)
		if !a.incoming {
			w.Write([]byte(a.str))
		}
	})
}
