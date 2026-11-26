package controller

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/bldsoft/gost/log"
)

const ArchivedQueryName = "archived"

const DefaultMaxBodyBytes int64 = 200 << 20

type BaseController struct {
	MaxBodyBytes int64
}

func (c BaseController) ResponseError(w http.ResponseWriter, err string, code int) {
	ResponseError(w, err, code)
}

func (c BaseController) ResponseOK(w http.ResponseWriter) {
	_, _ = w.Write([]byte("OK"))
}

func (c BaseController) ResponseJson(w http.ResponseWriter, r *http.Request, v any, needObjectLog ...bool) {
	w.Header().Set("Content-Type", "application/json")
	// err := json.NewEncoder(w).Encode(v)
	body, err := json.Marshal(v)
	if len(needObjectLog) == 0 || needObjectLog[0] {
		log.FromContext(r.Context()).TraceOrErrorWithFields(err, log.Fields{"body": string(body), "object": v}, "Encode response json")
	}
	_, err = w.Write(body)
	if err != nil {
		panic(err)
	}
}

func (c BaseController) GetObjectFromBody(w http.ResponseWriter, r *http.Request, obj any, needObjectLog ...bool) bool {
	var bodyBytes []byte
	if r.Body != nil {
		maxBytes := c.MaxBodyBytes
		if maxBytes <= 0 {
			maxBytes = DefaultMaxBodyBytes
		}
		if r.ContentLength > maxBytes {
			_ = r.Body.Close()
			c.ResponseError(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)

			return false
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
		var err error
		bodyBytes, err = io.ReadAll(r.Body)
		_ = r.Body.Close()
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			c.ResponseError(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)

			return false
		}
		if err != nil {
			c.ResponseError(w, err.Error(), http.StatusBadRequest)

			return false
		}
	}

	err := json.Unmarshal(bodyBytes, obj)
	if len(needObjectLog) == 0 || needObjectLog[0] {
		log.FromContext(r.Context()).TraceOrErrorWithFields(err, log.Fields{"body": string(bodyBytes), "object": obj}, "Decode request json")
	}
	if err != nil {
		c.ResponseError(w, err.Error(), http.StatusBadRequest)

		return false
	}

	return true
}
