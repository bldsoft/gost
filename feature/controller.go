package feature

import (
	"errors"
	"net/http"

	"github.com/bldsoft/gost/auth"
	"github.com/bldsoft/gost/config/feature"
	"github.com/bldsoft/gost/controller"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/utils"
	"github.com/go-chi/chi/v5"
)

type Controller struct {
	controller.BaseController
	featureService IFeatureService
}

func NewController(featureService IFeatureService) *Controller {
	return &Controller{featureService: featureService}
}

func (c *Controller) responseError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
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

// GetFeatureHandler get all features
// @Summary get all features
// @Tags admin
// @Security ApiKeyAuth
// @Produce text/yaml
// @Success 200 {array} Feature "OK"
// @Router /env/feature [get]
func (c *Controller) GetFeaturesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	features, err := c.featureService.GetAll(ctx)
	if err != nil {
		c.responseError(w, r, err)
		return
	}
	c.ResponseJson(w, r, features)
}

// GetFeatureHandler gets a single feature flag.
// @Summary gets a single feature flag.
// @Tags admin
// @Security ApiKeyAuth
// @Param id path string true "Feature name"
// @Produce text/yaml
// @Success 200 {object} Feature "OK"
// @Failure 404 {string} string "Not found"
// @Router /env/feature/{id} [get]
func (c *Controller) GetFeatureHandler(w http.ResponseWriter, r *http.Request) {
	id := feature.IdFromString(chi.URLParam(r, "id"))
	ctx := r.Context()
	feature, err := c.featureService.Get(ctx, id)
	if err != nil {
		c.responseError(w, r, err)
		return
	}
	c.ResponseJson(w, r, feature)
}

// PutFeatureHandler updates a feature flag.
// @Summary updates a feature flag.
// @Tags admin
// @Security ApiKeyAuth
// @Param id path string true "Feature name"
// @Consume json
// @Param Feature body Feature true "Feature"
// @Produce json, text/plain
// @Success 200 {object} Feature "OK"
// @Failure 400 {string} string "bad request"
// @Failure 404 {string} string "Not found"
// @Router /env/feature/{id} [patch]
func (c *Controller) PatchFeatureHandler(w http.ResponseWriter, r *http.Request) {
	var f *Feature
	if !c.GetObjectFromBody(w, r, &f) {
		return
	}
	ctx := r.Context()
	f.ID = feature.IdFromString(chi.URLParam(r, "id"))
	f, err := c.featureService.Update(ctx, f)
	if err != nil {
		c.responseError(w, r, err)
		return
	}
	c.ResponseJson(w, r, f)
}

func (c *Controller) Mount(r chi.Router) {
	r.Get("/", c.GetFeaturesHandler)
	r.Get("/{id}", c.GetFeatureHandler)
	r.Patch("/{id}", c.PatchFeatureHandler)
}
