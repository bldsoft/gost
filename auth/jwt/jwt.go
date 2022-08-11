package jwt

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth"
	"github.com/lestrrat-go/jwx/jwk"
)

type JwtConfig struct {
	Alg       string `mapstructure:"JWT_ALG" description:"Algorithm used to sign JWT"`
	PemPath   string `mapstructure:"JWT_PUBLIC_KEY_PATH" description:"Path to PEM encoded ASN.1 DER format key"`
	publicKey jwk.Key
}

func (c *JwtConfig) PublicKey() jwk.Key {
	return c.publicKey
}

func (c *JwtConfig) SetDefaults() {}

func (c *JwtConfig) Validate() (err error) {
	if len(c.PemPath) != 0 {
		bytes, err := ioutil.ReadFile(c.PemPath)
		if err != nil {
			return fmt.Errorf("failed to read jwt public key: %w", err)
		}

		c.publicKey, err = jwk.ParseKey(bytes, jwk.WithPEM(true))
		if err != nil {
			return fmt.Errorf("failed to parse jwt public key: %w", err)
		}
	}
	return nil
}

// JwtAuthMiddleware accepts either a raw key (e.g. rsa.PrivateKey, ecdsa.PrivateKey, etc)
// or a jwk.Key, and the name of the algorithm that should be used to sign the token.
func JwtAuthMiddlewareFromConfig(cfg JwtConfig) func(next http.Handler) http.Handler {
	return chi.Chain(jwtauth.Verifier(jwtauth.New(cfg.Alg, cfg.publicKey, nil)), jwtauth.Authenticator).Handler
}

// JwtAuthMiddleware accepts either a raw key (e.g. rsa.PrivateKey, ecdsa.PrivateKey, etc)
// or a jwk.Key, and the name of the algorithm that should be used to sign the token.
func JwtAuthMiddleware(alg string, signKey interface{}) func(next http.Handler) http.Handler {
	return chi.Chain(jwtauth.Verifier(jwtauth.New(alg, signKey, nil)), jwtauth.Authenticator).Handler
}
