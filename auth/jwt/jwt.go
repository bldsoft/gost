package jwt

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth"
)

// JwtAuthMiddleware accepts either a raw key (e.g. rsa.PrivateKey, ecdsa.PrivateKey, etc)
// or a jwk.Key, and the name of the algorithm that should be used to sign the token.
func JwtAuthMiddleware(alg string, signKey interface{}) func(next http.Handler) http.Handler {
	return chi.Chain(jwtauth.Verifier(jwtauth.New(alg, signKey, nil)), jwtauth.Authenticator).Handler
}
