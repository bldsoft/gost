package auth

import (
	"net/http"

	"github.com/bldsoft/gost/controller"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/utils"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
)

const SessionUserKey = "user"

// AuthController ...
type AuthController[PT AuthenticatablePtr[T], T any] struct {
	controller.BaseController
	sessionStore sessions.Store
	authService  IAuthService[PT, T]
	cookieName   string
}

func NewAuthController[PT AuthenticatablePtr[T], T any](sessionStore sessions.Store, service IAuthService[PT, T], cookieName string) *AuthController[PT, T] {
	return &AuthController[PT, T]{sessionStore: sessionStore, authService: service, cookieName: cookieName}
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

// AuthenticateMiddleware authenticates user and put it into into request context.
func (c *AuthController[PT, T]) AuthenticateMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, ok := c.session(w, r)
			if !ok {
				return
			}

			user, ok := session.Values[SessionUserKey].(T)
			if !ok {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, WithUserContext(r, &user))
		})

	}
}

// SignUp ...
func (c *AuthController[PT, T]) SignUp(w http.ResponseWriter, r *http.Request) {
	var user T
	if !c.GetObjectFromBody(w, r, &user) {
		return
	}

	if err := c.authService.CreateUser(r.Context(), &user); err != nil {
		c.ResponseError(w, err.Error(), http.StatusBadRequest)
		return
	}

	c.ResponseOK(w)
}

func (c *AuthController[PT, T]) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username, password, ok := r.BasicAuth()
	if !ok {
		c.ResponseError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	user, err := c.authService.Login(ctx, username, password)
	switch err {
	case utils.ErrObjectNotFound:
		c.ResponseError(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	case nil:
		session, ok := c.session(w, r)
		if !ok {
			return
		}
		session.Values[SessionUserKey] = *user
		if c.saveSession(w, r, session) {
			c.ResponseOK(w)
		}
	default:
		log.FromContext(ctx).Infof(err.Error())
		c.ResponseError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	}
}

// Logout ...
func (c *AuthController[PT, T]) Logout(w http.ResponseWriter, r *http.Request) {
	session, ok := c.session(w, r)
	if !ok {
		return
	}

	session.Values[SessionUserKey] = nil
	session.Options.MaxAge = -1

	if c.saveSession(w, r, session) {
		c.ResponseOK(w)
	}
}

func (c *AuthController[PT, T]) Mount(r chi.Router) {
	r.Post("/signup", c.SignUp)
	r.Post("/login", c.Login)
	r.Post("/logout", c.Logout)
}
