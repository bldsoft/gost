package server

import (
	"net/http"
	"sync/atomic"

	"github.com/go-chi/chi/v5"
)

type probesController struct {
	ready atomic.Bool
}

func newProbesController() *probesController {
	return &probesController{}
}

func (c *probesController) SetReady(v bool) *probesController {
	c.ready.Store(v)

	return c
}

func (c *probesController) HealthHandler(w http.ResponseWriter, r *http.Request) {
	c.writeCode(w, http.StatusOK)
}

func (c *probesController) ReadyHandler(w http.ResponseWriter, r *http.Request) {
	if !c.ready.Load() {
		c.writeCode(w, http.StatusServiceUnavailable)

		return
	}
	c.writeCode(w, http.StatusOK)
}

func (c *probesController) StartupHandler(w http.ResponseWriter, r *http.Request) {
	c.writeCode(w, http.StatusOK)
}

func (c *probesController) writeCode(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
	_, _ = w.Write([]byte(http.StatusText(code)))
}

func (c *probesController) Mount(r chi.Router) {
	r.Get("/healthy", c.HealthHandler)
	r.Get("/ready", c.ReadyHandler)
	r.Get("/startup", c.StartupHandler)
}
