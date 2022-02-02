package feature

import (
	"errors"
	"net/http"

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
		log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err}, "Failed to get features")
		c.ResponseError(w, "", http.StatusInternalServerError)
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
	feature := c.featureService.Get(r.Context(), id)
	if feature == nil {
		c.ResponseError(w, "Not found", http.StatusNotFound)
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
	switch {
	case errors.Is(err, utils.ErrObjectNotFound):
		c.ResponseError(w, "Not found", http.StatusNotFound)
	case err != nil:
		log.FromContext(ctx).Errorf("Failed to update feature: %s", err.Error())
		c.ResponseError(w, err.Error(), http.StatusBadRequest)
	default:
		c.ResponseJson(w, r, f)
	}
}

func (c *Controller) Mount(r chi.Router) {
	r.Get("/", c.GetFeaturesHandler)
	r.Get("/{id}", c.GetFeatureHandler)
	r.Patch("/{id}", c.PatchFeatureHandler)
}
