package auth

import (
	"net/http"

	"github.com/bldsoft/gost/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth"
	"github.com/lestrrat-go/jwx/jwk"
)

type JwtConfig struct {
	Alg       string             `mapstructure:"JWT_ALG"`
	PublicKey config.HidenString `mapstructure:"JWT_PUBLIC_KEY"`

	key interface{}
}

// SetDefaults ...
func (c *JwtConfig) SetDefaults() {
	c.Alg = "RS256"
}

// Validate ...
func (c *JwtConfig) Validate() (err error) {
	c.key, err = jwk.ParseKey([]byte(c.PublicKey.String()))
	return err
}

func JwtAuthMiddleware(config *JwtConfig) func(next http.Handler) http.Handler {
	return chi.Chain(jwtauth.Verifier(jwtauth.New(config.Alg, config.key, nil)), jwtauth.Authenticator).Handler
}
