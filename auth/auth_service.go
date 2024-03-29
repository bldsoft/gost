package auth

import (
	"context"
	"fmt"

	"github.com/bldsoft/gost/repository"
)

var (
	ErrWrongPassword = fmt.Errorf("wrong password")
	ErrNotActive     = fmt.Errorf("user is not active")
)

// AuthService ...
type AuthService[PT AuthenticablePtr[T], T any] struct {
	userRep        IAuthRepository[PT]
	passwordHasher PasswordHasher
}

// NewAuthService ...
func NewAuthService[PT AuthenticablePtr[T], T any](rep IAuthRepository[PT], passwordHasher PasswordHasher) *AuthService[PT, T] {
	return &AuthService[PT, T]{userRep: rep, passwordHasher: passwordHasher}
}

func (s *AuthService[PT, T]) Login(ctx context.Context, username, password string, opts ...*repository.QueryOptions) (PT, error) {
	user, err := s.userRep.FindByLogin(ctx, username, &repository.QueryOptions{Archived: false})
	if err != nil {
		return nil, err
	}

	if err := user.Active(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNotActive, err)
	}

	if err := s.passwordHasher.VerifyPassword(user.Password(), password); err != nil {
		return nil, ErrWrongPassword
	}
	return user, nil
}
