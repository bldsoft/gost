package controller

import (
	"fmt"
	"net/http"
)

type version struct {
	Version string `json:"version"`
} //@name Version

var Version string
var Commit string

func ProgramVersion() string {
	return fmt.Sprintf("%s-%s", Version, Commit)
}

type CommonController struct {
	BaseController
}

func NewCommonController() *CommonController {
	return &CommonController{}
}

// GetDefaultVersionHandler reponses with a service version.
// You can set Version field in code explicidly or by adding -ldflags to go build command:
// go build -ldflags="-X 'github.com/bldsoft/gost/controller.version=$(git describe --always)'"
// @Summary get service version
// @Tags public
// @Produce json
// @Success 200 {object} controller.version
// @Router /version [get]
func (c *CommonController) GetVersionHandler(w http.ResponseWriter, r *http.Request) {
	c.ResponseJson(w, r, version{ProgramVersion()})
}

// @Summary ping request
// @Tags public
// @Produce plain
// @Success 200 {string} string "OK"
// @Router /ping [get]
func (c *CommonController) GetPingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}
