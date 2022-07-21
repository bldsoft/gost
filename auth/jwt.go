package auth

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth"
)

func JwtAuthMiddleware(alg string, signKey interface{}) func(next http.Handler) http.Handler {
	return chi.Chain(jwtauth.Verifier(jwtauth.New(alg, signKey, nil)), jwtauth.Authenticator).Handler
}
