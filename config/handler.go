package config

import "net/http"

// GetEnvHandler get current environment
// @Summary get current environment
// @Tags admin
// @Security ApiKeyAuth
// @Produce text/yaml
// @Success 200 {string} string "OK"
// @Router /env [get]
func GetEnvHandler(config IConfig) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(FormatEnv(config)))
	}
}
