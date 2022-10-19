package log

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

type testController struct {
	Code      int
	Size      uint32
	ErrorSize int
}

func (c *testController) Handler(w http.ResponseWriter, r *http.Request) {
	if c.ErrorSize > 0 {
		ek := GetLogEntryFromRequest(r).(ErrorKeeper)
		ek.SetError(string(make([]byte, c.ErrorSize)))
	}

	w.Write(make([]byte, int(c.Size)))
	w.WriteHeader(c.Code)
}

func TestChannelFormatter(t *testing.T) {
	router := chi.NewRouter()
	requestC := make(chan *RequestInfo, 1)
	router.Use(ChanRequestLogger(requestC, ""))
	var c testController
	router.HandleFunc("/*", c.Handler)

	testCases := []struct {
		Method    string
		Path      string
		Size      uint32
		Code      int
		ErrorSize int
	}{
		{"GET", "/", 0, http.StatusOK, 0},
		{"POST", "/", 0, http.StatusOK, 0},
		{"PUT", "/", 0, http.StatusOK, 0},
		{"DELETE", "/", 0, http.StatusOK, 0},

		{"GET", "/a", 0, http.StatusOK, 0},
		{"GET", "/a/b", 0, http.StatusOK, 0},
		{"GET", "/a/b?query=123", 0, http.StatusOK, 0},

		{"GET", "/", 0, http.StatusOK, 0},
		{"GET", "/", 1, http.StatusOK, 0},
		{"GET", "/", 10, http.StatusOK, 0},
		{"GET", "/", 100000, http.StatusOK, 0},

		{"GET", "/", 0, http.StatusOK, 0},
		{"GET", "/", 0, http.StatusInternalServerError, 0},
		{"GET", "/", 0, http.StatusUnauthorized, 0},

		{"GET", "/", 0, http.StatusOK, 1},
		{"GET", "/", 0, http.StatusOK, 10},
		{"GET", "/", 0, http.StatusOK, 100},
	}

	for _, testCase := range testCases {
		rw := httptest.NewRecorder()
		req := httptest.NewRequest(testCase.Method, testCase.Path, nil)
		req.RemoteAddr = "127.0.0.1"
		userAgent := "agent"
		req.Header.Add("User-Agent", userAgent)
		c.Size, c.Code, c.ErrorSize = testCase.Size, testCase.Code, testCase.ErrorSize
		router.ServeHTTP(rw, req)
		requestInfo := <-requestC
		assert.Equal(t, req.RemoteAddr, requestInfo.ClientIp)
		assert.Equal(t, userAgent, requestInfo.UserAgent)
		assert.Equal(t, GetRequestMethodType(req.Method), requestInfo.RequestMethod)
		url, _ := url.Parse(requestInfo.Path)
		assert.Equal(t, req.URL.Path, url.Path)
		assert.Equal(t, len(rw.Body.Bytes()), int(requestInfo.Size))
		assert.Equal(t, rw.Code, int(requestInfo.ResponseCode))
		assert.Equal(t, testCase.ErrorSize, len(requestInfo.Error))
	}
}

func TestChannelFormatterRequestError(t *testing.T) {
	router := chi.NewRouter()
	requestC := make(chan *RequestInfo, 1)
	router.Use(ChanRequestLogger(requestC, ""))

	router.Use(func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ek := GetLogEntryFromRequest(r).(ErrorKeeper)
			ek.SetError("1")

			next.ServeHTTP(w, r)
			ek.SetError(ek.Error() + "3")
		}
		return http.HandlerFunc(fn)
	})

	router.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		ek := GetLogEntryFromRequest(r).(ErrorKeeper)
		ek.SetError(ek.Error() + "2")
	})

	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	router.ServeHTTP(rw, req)
	requestInfo := <-requestC
	assert.Equal(t, "123", requestInfo.Error)
}

func TestChannelFormatterCustomErr(t *testing.T) {
	type CustomRequestInfo struct {
		RequestInfo
		CustomField string
	}

	router := chi.NewRouter()
	requestC := make(chan *CustomRequestInfo, 1)
	router.Use(ChanRequestLogger(requestC, ""))

	router.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		requestInfo := r.Context().Value(RequestInfoCtxKey).(*CustomRequestInfo)
		requestInfo.CustomField = "some value"
	})

	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	router.ServeHTTP(rw, req)
	requestInfo := <-requestC

	assert.Equal(t, requestInfo.CustomField, "some value")
}
