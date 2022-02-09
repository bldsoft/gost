package auth

import (
	"context"
	"net/http"

	"github.com/bldsoft/gost/log"
	"golang.org/x/exp/slices"
)

var UserEntryCtxKey interface{} = "UserEntry"

// WithUserContext sets the User entry for a request.
func WithUserContext[T any](r *http.Request, entry T) *http.Request {
	r = r.WithContext(context.WithValue(r.Context(), UserEntryCtxKey, entry))
	return r
}

// FromContext returns the User entry for a request.
func FromContext[T any](ctx context.Context, allowPanic bool) (T, bool) {
	entry, ok := ctx.Value(UserEntryCtxKey).(T)
	if !ok && allowPanic {
		log.FromContext(ctx).Panic("User not found")
	}
	return entry, ok
}

func AuthorizationMiddleware[T IRole](roles ...T) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := FromContext[Authorizable[T]](r.Context(), false)
			switch {
			case !ok:
				log.FromContext(r.Context()).Warn("User authorization failed")
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			case !slices.Contains(roles, user.Role()):
				log.FromContext(r.Context()).WarnWithFields(log.Fields{"user": user}, "Forbidden")
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			default:
				next.ServeHTTP(w, r)
			}
		})
	}
}
