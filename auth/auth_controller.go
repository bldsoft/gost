package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/bldsoft/gost/controller"
	"github.com/bldsoft/gost/log"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
)

const SessionUserKey = "user"

var (
	UserEntryCtxKey    interface{} = "UserEntry"
	SessionEntryCtxKey interface{} = "SessionEntry"
)

func requestContextAdd(r *http.Request, key, value interface{}) *http.Request {
	r = r.WithContext(context.WithValue(r.Context(), key, value))
	return r
}

// WithUserContext sets the User entry for a request.
func WithUserContext[T any](r *http.Request, user T) *http.Request {
	return requestContextAdd(r, UserEntryCtxKey, user)
}

func withSessionRequest(r *http.Request, s *sessions.Session) *http.Request {
	return requestContextAdd(r, SessionEntryCtxKey, wrapSession(s))
}

func withSessionContext(ctx context.Context, s *sessions.Session) context.Context {
	return context.WithValue(ctx, SessionEntryCtxKey, wrapSession(s))
}

// UserFromContext returns User session
func SessionFromContext(ctx context.Context) *Session {
	s, _ := ctx.Value(SessionEntryCtxKey).(*Session)
	return s
}

// UserFromContext returns the User entry for a request.
func UserFromContext(ctx context.Context) interface{} {
	return ctx.Value(UserEntryCtxKey)
}

// AuthController ...
type AuthController[PT AuthenticablePtr[T], T any] struct {
	controller.BaseController
	authService  IAuthService[PT, T]
	sessionStore sessions.Store
	cookieName   string
}

func NewAuthController[PT AuthenticablePtr[T], T any](service IAuthService[PT, T], sessionStore sessions.Store, cookieName string) *AuthController[PT, T] {
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

func (c *AuthController[PT, T]) deleteSession(w http.ResponseWriter, r *http.Request, session *sessions.Session) bool {
	// Delete session (MaxAge <= 0)
	session.Options.MaxAge = -1
	return c.saveSession(w, r, session)
}

// AuthenticateMiddleware authenticates user and put it into request context.
func (c *AuthController[PT, T]) AuthenticateMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := c.sessionStore.Get(r, c.cookieName)
			if err != nil {
				log.FromContext(r.Context()).ErrorWithFields(log.Fields{"err": err}, "bad session")
				http.SetCookie(w, &http.Cookie{Name: c.cookieName, MaxAge: -1, Path: "/"})
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			user, ok := session.Values[SessionUserKey].(T)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, WithUserContext(withSessionRequest(r, session), &user))
			c.saveSession(w, r, session)
		})
	}
}

func (c *AuthController[PT, T]) Login(w http.ResponseWriter, r *http.Request) {
	var creds T
	if !c.BaseController.GetObjectFromBody(w, r, &creds, false) {
		return
	}

	session, ok := c.session(w, r)
	if !ok {
		return
	}
	ctx := withSessionContext(r.Context(), session)

	user, err := c.authService.Login(ctx, PT(&creds).Login(), PT(&creds).Password())
	switch {
	case err == nil:
		session.Values[SessionUserKey] = *user
		if c.saveSession(w, r, session) {
			// log.FromContext(ctx).InfoWithFields(log.Fields{"login": user.Login()}, "User is logged in")
			c.ResponseJson(w, r, user)
		}
	case errors.Is(err, ErrNotActive):
		c.ResponseError(w, err.Error(), http.StatusForbidden)
	default:
		log.FromContext(ctx).Error(err.Error())
		c.ResponseError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	}
}

// Refresh ...
func (c *AuthController[PT, T]) Refresh(w http.ResponseWriter, r *http.Request) {
	session, ok := c.session(w, r)
	if !ok {
		return
	}
	if c.saveSession(w, r, session) {
		c.ResponseOK(w)
	}
}

// Logout ...
func (c *AuthController[PT, T]) Logout(w http.ResponseWriter, r *http.Request) {
	session, ok := c.session(w, r)
	if !ok {
		return
	}
	if c.deleteSession(w, r, session) {
		c.ResponseOK(w)
	}
}

func (c *AuthController[PT, T]) Mount(r chi.Router) {
	r.Post("/login", c.Login)
	r.Post("/logout", c.Logout)
	r.Group(func(r chi.Router) {
		r.Use(c.AuthenticateMiddleware())
		r.Post("/refresh", c.Refresh)
	})
}
