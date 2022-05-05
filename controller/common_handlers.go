package controller

import (
	"net/http"

	"github.com/bldsoft/gost/config"
	v "github.com/bldsoft/gost/version"
	"github.com/ghodss/yaml"
)

type version struct {
	Version   string `json:"version,omitempty"`
	GitBranch string `json:"branch,omitempty"`
	GitCommit string `json:"commit,omitempty"`
} //@name Version

// GetDefaultVersionHandler reponses with a service version.
// You can set Version field in code explicidly or by adding -ldflags to go build command:
// go build -ldflags="-X 'github.com/bldsoft/gost/version.Version=1.0.0'"
// @Summary get service version
// @Tags public
// @Produce json
// @Success 200 {object} controller.version
// @Router /version [get]
func GetVersionHandler(w http.ResponseWriter, r *http.Request) {
	BaseController{}.ResponseJson(w, r, version{
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
func GetPingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

// GetEnvHandler get current environment
// @Summary get current environment
// @Tags admin
// @Security ApiKeyAuth
// @Produce text/yaml
// @Success 200 {string} string "OK"
// @Router /env [get]
func GetEnvHandler(cfg config.IConfig, features interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(config.FormatEnv(cfg)))

		yamlFeatures, _ := yaml.Marshal(struct {
			Features interface{}
		}{
			Features: features,
		})
		w.Write(([]byte(yamlFeatures)))
	}
}
