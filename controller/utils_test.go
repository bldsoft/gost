package controller

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bldsoft/gost/utils"
	"github.com/stretchr/testify/assert"
)

type test interface {
	name() string
	check(t *testing.T)
}

type getQueryTest[T utils.Parsed] struct {
	queryStr, paramName       string
	expectedRes, defaultValue T
	errExpected               bool
}

func newTest[T utils.Parsed](query, paramName string, expectedRes, defaultValue T, errExpected bool) *getQueryTest[T] {
	return &getQueryTest[T]{query, paramName, expectedRes, defaultValue, errExpected}
}

func (test *getQueryTest[T]) check(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example/api?"+test.queryStr, nil)
	assert.NoError(t, err)
	var res, zero T
	if test.defaultValue == zero {
		res, err = GetQueryOption[T](req, test.paramName)
	} else {
		res, err = GetQueryOption(req, test.paramName, test.defaultValue)
	}

	if test.errExpected {
		assert.Error(t, err)
	} else {
		assert.NoError(t, err)
		assert.Equal(t, test.expectedRes, res)
	}

	// ParseQueryOption
	w := httptest.NewRecorder()
	outRes := test.defaultValue
	assert.Equal(t, !test.errExpected, ParseQueryOption(req, w, test.paramName, &outRes))
	if test.errExpected {
		assert.Equal(t, w.Code, http.StatusBadRequest)
		return
	}
	assert.Equal(t, test.expectedRes, outRes)
}

func (test *getQueryTest[T]) name() string {
	if test.errExpected {
		return fmt.Sprintf("uri=%s expected error", test.queryStr)
	}
	return fmt.Sprintf("%s expected=%v", test.queryStr, test.expectedRes)
}

func TestGetAndParseQueryOption(t *testing.T) {
	tests := []test{
		newTest("archived=true", "archived", true, false, false),
		newTest("archived=false", "archived", false, false, false),
		newTest("", "archived", false, false, false),
		newTest("", "archived", true, true, false),
		newTest("archived=123", "archived", true, false, true),

		newTest("intParam=123", "intParam", 123, 0, false),
		newTest("intParam=123", "intParam", 123, 1, false),

		newTest("stringParam=123", "stringParam", "123", "", false),
		newTest("stringParam=123", "stringParam", "123", "1", false),
	}

	for _, test := range tests {
		t.Run(test.name(), test.check)
	}
}

func TestGetQueryOptionSlice(t *testing.T) {
	tests := []struct {
		queryStr    string
		expectedRes []int
		errExpected bool
	}{
		{"intParams=1", []int{1}, false},
		{"intParams=1,2,3", []int{1, 2, 3}, false},
		{"", nil, false},
		{"intParams=1,2,asdf", nil, true},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			req, err := http.NewRequest("GET", "http://example/api?"+test.queryStr, nil)
			assert.NoError(t, err)

			res, err := GetQueryOptionSlice[int](req, "intParams")
			if test.errExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err, "unexpected err")
				assert.Equal(t, test.expectedRes, res)
			}
			//ParseQueryOptionSlice
			w := httptest.NewRecorder()
			var outRes []int
			assert.Equal(t, !test.errExpected, ParseQueryOptionSlice(req, w, "intParams", &outRes))
			if test.errExpected {
				assert.Equal(t, w.Code, http.StatusBadRequest)
				assert.Equal(t, test.expectedRes, outRes)
				return
			}
			assert.Equal(t, test.expectedRes, outRes)
		})
	}
}
