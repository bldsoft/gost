package jwt

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
)

type JwtConfig struct {
	key     jwk.Key
	Alg     string `mapstructure:"JWT_ALG" description:"Algorithm used to sign JWT"`
	PemPath string `mapstructure:"JWT_KEY_PATH" description:"Path to PEM encoded ASN.1 DER format key"`
}

func (c *JwtConfig) PrivateKey() jwk.Key {
	return c.key
}

func (c *JwtConfig) PublicKey() jwk.Key {
	public, _ := c.key.PublicKey()
	return public
}

func (c *JwtConfig) SetDefaults() {}

func (c *JwtConfig) Validate() (err error) {
	if len(c.PemPath) != 0 {
		bytes, err := ioutil.ReadFile(c.PemPath)
		if err != nil {
			return fmt.Errorf("failed to read jwt key: %w", err)
		}

		c.key, err = jwk.ParseKey(bytes, jwk.WithPEM(true))

		if err != nil {
			return fmt.Errorf("failed to parse jwt key: %w", err)
		}
	}
	return nil
}

func JwtAuthMiddlewareFromConfig(cfg JwtConfig) func(next http.Handler) http.Handler {
	return JwtAuthMiddleware(cfg.Alg, cfg.PublicKey())
}

// JwtAuthMiddleware accepts either a raw key (e.g. rsa.PrivateKey, ecdsa.PrivateKey, etc)
// or a jwk.Key, and the name of the algorithm that should be used to sign the token.
func JwtAuthMiddleware(alg string, signKey interface{}) func(next http.Handler) http.Handler {
	return chi.Chain(jwtauth.Verifier(jwtauth.New(alg, signKey, nil)), jwtauth.Authenticator).Handler
}

func SignToken(token jwt.Token, cfg JwtConfig) (signed []byte, err error) {
	return jwt.Sign(token, jwa.SignatureAlgorithm(cfg.Alg), cfg.PrivateKey())
}
