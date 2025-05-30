package controller

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/bldsoft/gost/log"
)

const ArchivedQueryName = "archived"

type BaseController struct{}

func (c BaseController) ResponseError(w http.ResponseWriter, err string, code int) {
	ResponseError(w, err, code)
}

func (c BaseController) ResponseOK(w http.ResponseWriter) {
	w.Write([]byte("OK"))
}

func (c BaseController) ResponseJson(w http.ResponseWriter, r *http.Request, v interface{}, needObjectLog ...bool) {
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

func (c BaseController) GetObjectFromBody(w http.ResponseWriter, r *http.Request, obj interface{}, needObjectLog ...bool) bool {
	var bodyBytes []byte
	if r.Body != nil {
		if contentlen := r.ContentLength; contentlen <= 0 {
			bodyBytes, _ = ioutil.ReadAll(r.Body)
		} else {
			bodyBytes = make([]byte, contentlen)
			io.ReadFull(r.Body, bodyBytes)
		}
		r.Body.Close()
		// r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// err := json.NewDecoder(bytes.NewBuffer(bodyBytes)).Decode(obj)
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
