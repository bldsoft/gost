package stat

import (
	"context"
	"net/http"

	"github.com/bldsoft/gost/controller"
	"github.com/go-chi/chi/v5"
)

type IStatService interface {
	Stats(ctx context.Context) []Stat
}

type Controller struct {
	controller.BaseController
	statService IStatService
}

func NewController(statCollectors ...StatCollector) *Controller {
	return NewControllerFromService(NewService(statCollectors...))
}

func NewControllerFromService(service IStatService) *Controller {
	return &Controller{
		statService: service,
	}
}

func (c Controller) StatHandler(w http.ResponseWriter, r *http.Request) {
	healthChecks := c.statService.Stats(r.Context())
	c.ResponseJson(w, r, healthChecks)
}

func (c *Controller) Mount(r chi.Router) {
	r.Get("/", c.StatHandler)
}
