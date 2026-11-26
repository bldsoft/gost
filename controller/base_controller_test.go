package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type bodyObject struct {
	Name string `json:"name"`
}

func TestGetObjectFromBodyValidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"ok"}`))
	w := httptest.NewRecorder()

	var obj bodyObject
	ok := BaseController{}.GetObjectFromBody(w, r, &obj)

	assert.True(t, ok)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, bodyObject{Name: "ok"}, obj)
}

func TestGetObjectFromBodyHugeContentLength(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"ok"}`))
	r.ContentLength = 10 << 30
	w := httptest.NewRecorder()

	var obj bodyObject
	ok := BaseController{}.GetObjectFromBody(w, r, &obj)

	assert.False(t, ok)
	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	assert.Equal(t, bodyObject{}, obj)
}

func TestGetObjectFromBodyChunkedTooLarge(t *testing.T) {
	const maxBytes int64 = 8
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"too-long"}`))
	r.ContentLength = -1
	w := httptest.NewRecorder()

	var obj bodyObject
	ok := BaseController{MaxBodyBytes: maxBytes}.GetObjectFromBody(w, r, &obj)

	assert.False(t, ok)
	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	assert.Equal(t, bodyObject{}, obj)
}

func TestGetObjectFromBodyInvalidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{`))
	w := httptest.NewRecorder()

	var obj bodyObject
	ok := BaseController{}.GetObjectFromBody(w, r, &obj)

	assert.False(t, ok)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, bodyObject{}, obj)
}
