package test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bldsoft/gost/routing"
	"github.com/stretchr/testify/assert"
)

func TestRouting(t *testing.T) {
	type args struct {
		r       *http.Request
		rule    routing.Rule
		handler http.HandlerFunc
	}
	tests := []struct {
		name string
		args args
		want *http.Response
	}{
		{
			name: "handle without conditions",
			args: args{
				r:    httptest.NewRequest(http.MethodGet, "http://example.com", nil),
				rule: routing.NewRule(routing.NoCondition, routing.HandleAction),
				handler: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}},
			want: &http.Response{
				StatusCode: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := routing.Routing(tt.args.rule)(tt.args.handler)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, tt.args.r)
			ResponseEqual(t, tt.want, w.Result())
		})
	}
}

func ResponseEqual(t *testing.T, want, got *http.Response) {
	assert.Equal(t, want.StatusCode, got.StatusCode)

	if want.ContentLength > 0 {
		assert.Equal(t, want.ContentLength, got.ContentLength)
		wantBody := ResponseBody(t, want)
		if len(wantBody) > 0 {
			assert.Equal(t, wantBody, ResponseBody(t, got))
		}
	}

	if len(want.Header) > 0 {
		assert.Equal(t, want.Header, got.Header)
	}
}

func ResponseBody(t *testing.T, resp *http.Response) []byte {
	if resp.ContentLength == 0 {
		return nil
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	return data
}
