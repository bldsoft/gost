package auth

import (
	"context"
	"fmt"

	"github.com/bldsoft/gost/repository"
)

var ErrWrongPassword = fmt.Errorf("wrong password")

// AuthService ...
type AuthService[PT AuthenticablePtr[T], T any] struct {
	userRep        IAuthRepository[PT]
	passwordHasher PasswordHasher
}

// NewAuthService ...
func NewAuthService[PT AuthenticablePtr[T], T any](rep IAuthRepository[PT], passwordHasher PasswordHasher) *AuthService[PT, T] {
	return &AuthService[PT, T]{userRep: rep, passwordHasher: passwordHasher}
}

func (s *AuthService[PT, T]) Login(ctx context.Context, username, password string) (PT, error) {
	user, err := s.userRep.FindByLogin(ctx, username, &repository.QueryOptions{Archived: false})
	if err != nil {
		return nil, err
	}

	if err := s.passwordHasher.VerifyPassword(user.Password(), password); err != nil {
		return nil, ErrWrongPassword
	}
	return user, nil
}
