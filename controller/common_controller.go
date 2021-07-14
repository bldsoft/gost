package controller

import (
	"net/http"
)

type Version struct {
	Version string `json:"version"`
}

var version string

type CommonController struct {
	BaseController
	Version Version
}

func NewCommonController() *CommonController {
	v := Version{Version: version}
	return &CommonController{Version: v}
}

func (c *CommonController) GetVersionHandler(version interface{}) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c.ResponseJson(w, r, version)
	}
}

// GetDefaultVersionHandler reponses with a service version.
// You can set Version field in code explicidly or by adding -ldflags to go build command:
// go build -ldflags="-X 'github.com/bldsoft/gost/controller.version=$(git describe --always)'"
// @Summary get service version
// @Tags public
// @Produce json
// @Success 200 {object} controller.Version
// @Router /version [get]
func (c *CommonController) GetDefaultVersionHandler() func(http.ResponseWriter, *http.Request) {
	return c.GetVersionHandler(c.Version)
}

// @Summary ping request
// @Tags public
// @Produce plain
// @Success 200 {string} string "OK"
// @Router /ping [get]
func (c *CommonController) GetPingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}
