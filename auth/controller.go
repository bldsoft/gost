package auth

import (
	"net/http"

	"github.com/bldsoft/gost/controller"
	"github.com/bldsoft/gost/log"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
)

const SessionUserKey = "user"

// AuthController ...
type AuthController[PT AuthenticatablePtr[T], T any] struct {
	controller.BaseController
	authService  IAuthService[PT, T]
	sessionStore sessions.Store
	cookieName   string
}

func NewAuthController[PT AuthenticatablePtr[T], T any](rep IAuthRepository[PT], hasher PasswordHasher, sessionStore sessions.Store, cookieName string) *AuthController[PT, T] {
	service := NewAuthService[PT, T](rep, hasher)
	return &AuthController[PT, T]{authService: service, sessionStore: sessionStore, cookieName: cookieName}
}

func (c *AuthController[PT, T]) Service() IAuthService[PT, T] {
	return c.authService
}

func (c *AuthController[PT, T]) session(w http.ResponseWriter, r *http.Request) (*sessions.Session, bool) {
	session, err := c.sessionStore.Get(r, c.cookieName)
	if err != nil {
		log.FromContext(r.Context()).ErrorWithFields(log.Fields{"err": err}, "Failed to get session")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return nil, false
	}
	return session, true
}

func (c *AuthController[PT, T]) saveSession(w http.ResponseWriter, r *http.Request, s *sessions.Session) bool {
	err := c.sessionStore.Save(r, w, s)
	if err != nil {
		log.FromContext(r.Context()).ErrorWithFields(log.Fields{"err": err}, "Failed to save session")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return false
	}
	return true
}

// AuthenticateMiddleware authenticates user and put it into request context.
func (c *AuthController[PT, T]) AuthenticateMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, ok := c.session(w, r)
			if !ok {
				return
			}

			user, ok := session.Values[SessionUserKey].(T)
			if !ok {
				log.FromContext(r.Context()).Error("User session isn't found")
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, WithUserContext(r, &user))
		})

	}
}

func (c *AuthController[PT, T]) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username, password, ok := r.BasicAuth()
	if !ok {
		log.FromContext(ctx).Info("Failed to get username and password")
		c.ResponseError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	user, err := c.authService.Login(ctx, username, password)
	switch err {
	case nil:
		session, ok := c.session(w, r)
		if !ok {
			return
		}
		session.Values[SessionUserKey] = *user
		if c.saveSession(w, r, session) {
			log.FromContext(ctx).InfoWithFields(log.Fields{"login": user.Login()}, "User is logged in")
			c.ResponseOK(w)
		}
	default:
		log.FromContext(ctx).Error(err.Error())
		c.ResponseError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	}
}

// Logout ...
func (c *AuthController[PT, T]) Logout(w http.ResponseWriter, r *http.Request) {
	session, ok := c.session(w, r)
	if !ok {
		return
	}

	// Delete session (MaxAge <= 0)
	session.Options.MaxAge = -1
	if c.saveSession(w, r, session) {
		c.ResponseOK(w)
	}
}

func (c *AuthController[PT, T]) Mount(r chi.Router) {
	r.Post("/login", c.Login)
	r.Post("/logout", c.Logout)
}
