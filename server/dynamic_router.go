package server

import (
	"net/http"
	"sync/atomic"
)

type dynamicRouter struct {
	atomic.Value
}

func (d *dynamicRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d.Load().(http.Handler).ServeHTTP(w, r)
}

func (d *dynamicRouter) Set(h http.Handler) {
	d.Store(http.HandlerFunc(h.ServeHTTP))
}
