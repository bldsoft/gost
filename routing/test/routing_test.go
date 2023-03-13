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
			name: "last outgoing redirect",
			args: args{
				r: httptest.NewRequest(http.MethodGet, "http://example.com", nil),
				rule: routing.NewRule(nil,
					routing.JoinActions(
						routing.ActionRedirect{IncomingRequest: false, Code: http.StatusMovedPermanently, Host: "outgoing1"},
						routing.ActionRedirect{IncomingRequest: false, Code: http.StatusMovedPermanently, Host: "outgoing2"},
						routing.ActionRedirect{IncomingRequest: false, Code: http.StatusMovedPermanently, Host: "outgoing3"},
					)),
				handler: OkHandler,
			},
			want: &http.Response{
				StatusCode: http.StatusMovedPermanently,
				Header: http.Header{
					"Location": []string{"http://outgoing3"},
				},
			},
		},
		{
			name: "last incoming redirect",
			args: args{
				r: httptest.NewRequest(http.MethodGet, "http://example.com", nil),
				rule: routing.NewRule(nil,
					routing.JoinActions(
						routing.ActionRedirect{IncomingRequest: true, Code: http.StatusMovedPermanently, Host: "incoming1"},
						routing.ActionRedirect{IncomingRequest: true, Code: http.StatusMovedPermanently, Host: "incoming2"},
						routing.ActionRedirect{IncomingRequest: true, Code: http.StatusMovedPermanently, Host: "incoming3"},
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
			name: "outgoing match, incoming action",
			args: args{
				r: httptest.NewRequest(http.MethodGet, "http://example.com", nil),
				rule: routing.NewRule(
					routing.NewFieldCondition[int](routing.StatusCode, routing.AnyOf(http.StatusOK)),
					WriteBodyAction{incoming: true, str: "1"},
				),
				handler: OkHandler,
			},
			want: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("OK")), // incoming action doesn't affect response if it matches after handling
			},
		},
		{
			name: "incoming match any success, incoming action",
			args: args{
				r: httptest.NewRequest(http.MethodGet, "http://example.com", nil),
				rule: routing.NewRule(
					routing.JoinConditions(false,
						routing.NewFieldCondition[int](routing.StatusCode, routing.AnyOf(http.StatusOK)),
						routing.NewFieldCondition[string](routing.Host, routing.AnyOf("example.com")),
					),
					WriteBodyAction{incoming: true, str: "1"},
				),
				handler: OkHandler,
			},
			want: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("1OK")), // managed to match before handling
			},
		},
		{
			name: "incoming match any fail, incoming action",
			args: args{
				r: httptest.NewRequest(http.MethodGet, "http://example.com", nil),
				rule: routing.NewRule(
					routing.JoinConditions(false,
						routing.NewFieldCondition[int](routing.StatusCode, routing.AnyOf(http.StatusOK)),
						routing.NewFieldCondition[string](routing.Host, routing.AnyOf("not-example.com")),
					),
					WriteBodyAction{incoming: true, str: "1"},
				),
				handler: OkHandler,
			},
			want: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("OK")), // incoming action doesn't affect response if it matches after handling
			},
		},
		{
			name: "outgoing match any success, outgoing action",
			args: args{
				r: httptest.NewRequest(http.MethodGet, "http://example.com", nil),
				rule: routing.NewRule(
					routing.JoinConditions(false,
						routing.NewFieldCondition[int](routing.StatusCode, routing.AnyOf(http.StatusOK)),
						routing.NewFieldCondition[string](routing.Host, routing.AnyOf("not-example.com")),
					),
					WriteBodyAction{incoming: false, str: "1"},
				),
				handler: OkHandler,
			},
			want: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("OK1")),
			},
		},
		{
			name: "incoming match all success, incoming action",
			args: args{
				r: httptest.NewRequest(http.MethodGet, "http://example.com", nil),
				rule: routing.NewRule(
					routing.JoinConditions(true,
						routing.NewFieldCondition[string](routing.Host, routing.AnyOf("example.com")),
						routing.NewFieldCondition[string](routing.Path, routing.AnyOf("")),
					),
					WriteBodyAction{incoming: true, str: "1"},
				),
				handler: OkHandler,
			},
			want: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("1OK")), // managed to match before handling
			},
		},
		{
			name: "incoming match all fail, incoming action",
			args: args{
				r: httptest.NewRequest(http.MethodGet, "http://example.com", nil),
				rule: routing.NewRule(
					routing.JoinConditions(true,
						routing.NewFieldCondition[string](routing.Host, routing.AnyOf("not-example.com")),
						routing.NewFieldCondition[string](routing.Path, routing.AnyOf("")),
					),
					WriteBodyAction{incoming: true, str: "1"},
				),
				handler: OkHandler,
			},
			want: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("OK")), // incoming action doesn't affect response if it matches after handling
			},
		},
		{
			name: "outgoing match all success, outgoing action",
			args: args{
				r: httptest.NewRequest(http.MethodGet, "http://example.com", nil),
				rule: routing.NewRule(
					routing.JoinConditions(true,
						routing.NewFieldCondition[string](routing.Host, routing.AnyOf("example.com")),
						routing.NewFieldCondition[int](routing.StatusCode, routing.AnyOf(http.StatusOK)),
					),
					WriteBodyAction{incoming: false, str: "1"},
				),
				handler: OkHandler,
			},
			want: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("OK1")),
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

func EmptyBodyOkHandler(t *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
}

func ResponseEqual(t *testing.T, want, got *http.Response) {
	assert.Equal(t, want.StatusCode, got.StatusCode)

	if want.Body != nil {
		wantBody := ResponseBody(t, want)
		assert.Equal(t, string(wantBody), string(ResponseBody(t, got)))
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

func (a WriteBodyAction) DoBeforeHandle(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request, error) {
	if a.incoming {
		_, _ = w.Write([]byte(a.str))
	}
	return w, r, nil
}

func (a WriteBodyAction) DoAfterHandle(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request, error) {
	if !a.incoming {
		_, _ = w.Write([]byte(a.str))
	}
	return w, r, nil
}
