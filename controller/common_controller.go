package controller

import (
	"net/http"

	v "github.com/bldsoft/gost/version"
)

type version struct {
	Version   string `json:"version,omitempty"`
	GitBranch string `json:"branch,omitempty"`
	GitCommit string `json:"commit,omitempty"`
} //@name Version

type CommonController struct {
	BaseController
}

func NewCommonController() *CommonController {
	return &CommonController{}
}

// GetDefaultVersionHandler reponses with a service version.
// You can set Version field in code explicidly or by adding -ldflags to go build command:
// go build -ldflags="-X 'github.com/bldsoft/gost/version.Version=1.0.0'"
// @Summary get service version
// @Tags public
// @Produce json
// @Success 200 {object} controller.version
// @Router /version [get]
func (c *CommonController) GetVersionHandler(w http.ResponseWriter, r *http.Request) {
	c.ResponseJson(w, r, version{
		Version:   v.Version,
		GitBranch: v.GitBranch,
		GitCommit: v.GitCommit,
	})
}

// @Summary ping request
// @Tags public
// @Produce plain
// @Success 200 {string} string "OK"
// @Router /ping [get]
func (c *CommonController) GetPingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}
