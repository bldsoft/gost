package changelog

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/bldsoft/gost/controller"
	"github.com/go-chi/chi/v5"
)

type Controller struct {
	controller.BaseController
	changeLogService IChangeLogService
}

func NewController(rep IChangeLogRepository) *Controller {
	return NewControllerByService(NewService(rep))
}

func NewControllerByService(changeLogService IChangeLogService) *Controller {
	return &Controller{changeLogService: changeLogService}
}

func (c *Controller) getFilter(r *http.Request) *Filter {
	query := r.URL.Query()
	var filter Filter
	if time, err := strconv.ParseInt(query.Get("startTime"), 10, 64); err == nil && time > 0 {
		filter.StartTime = time
	}
	if time, err := strconv.ParseInt(query.Get("endTime"), 10, 64); err == nil && time > filter.StartTime {
		filter.EndTime = time
	}
	if entityID := query.Get("entityID"); entityID != "" {
		filter.EntityID = entityID
	}
	if collections := query.Get("entities"); collections != "" {
		filter.Collections = strings.Split(collections, ",")
	}
	return &filter
}

func (c *Controller) GetHandler(w http.ResponseWriter, r *http.Request) {
	records, err := c.changeLogService.GetRecords(r.Context(), c.getFilter(r))
	if err != nil {
		c.ResponseError(w, err.Error(), http.StatusNotFound)
		return
	}
	c.ResponseJson(w, r, records)
}

func (c *Controller) Mount(r chi.Router) {
	r.Get("/", c.GetHandler)
}
