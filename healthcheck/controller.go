package healthcheck

import (
	"context"
	"net/http"

	"github.com/bldsoft/gost/controller"
	"github.com/go-chi/chi/v5"
)

type Service interface {
	CheckHealth(ctx context.Context) []Health
}

type Controller struct {
	controller.BaseController
	healthCheckService Service
}

func NewController(service Service) *Controller {
	return &Controller{
		healthCheckService: service,
	}
}

func (c Controller) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	healthChecks := c.healthCheckService.CheckHealth(r.Context())
	c.ResponseJson(w, r, healthChecks)
}

func (c *Controller) Mount(r chi.Router) {
	r.Get("/", c.HealthCheckHandler)
}
