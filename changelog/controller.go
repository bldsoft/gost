package changelog

import (
	"errors"
	"net/http"

	"github.com/bldsoft/gost/auth"
	"github.com/bldsoft/gost/controller"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/utils"
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

func (c *Controller) GetHandler(w http.ResponseWriter, r *http.Request) {
	params, err := utils.FromRequest[RecordsParams](r)
	if err != nil {
		c.ResponseError(w, err.Error(), http.StatusBadRequest)
		return
	}
	records, err := c.changeLogService.GetRecords(r.Context(), params)
	switch {
	case err == nil:
		c.ResponseJson(w, r, records)
	case errors.Is(err, utils.ErrObjectNotFound):
		c.ResponseError(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	case errors.Is(err, auth.ErrUnauthorized):
		c.ResponseError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	case errors.Is(err, auth.ErrForbidden):
		c.ResponseError(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
	default:
		log.FromContext(r.Context()).Error(err.Error())
		c.ResponseError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (c *Controller) Mount(r chi.Router) {
	r.Get("/", c.GetHandler)
}
