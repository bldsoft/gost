package auth

import (
	"context"
	"fmt"

	"github.com/bldsoft/gost/repository"
	"golang.org/x/crypto/bcrypt"
)

var ErrWrongPassword = fmt.Errorf("wrong password")

// AuthService ...
type AuthService[PT AuthenticatablePtr[T], T any] struct {
	userRep IUserRepository[PT]
}

// NewAuthService ...
func NewAuthService[PT AuthenticatablePtr[T], T any](rep IUserRepository[PT]) *AuthService[PT, T] {
	return &AuthService[PT, T]{userRep: rep}
}

func (s *AuthService[PT, T]) UserFromContext(ctx context.Context, allowPanic bool) (PT, bool) {
	return FromContext[PT](ctx, allowPanic)
}

func (s *AuthService[PT, T]) prepareEntity(user PT) error {
	hashedPass, err := s.HashAndSalt(user.Password())
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}
	user.SetPassword(hashedPass)
	return nil
}

func (s *AuthService[PT, T]) CreateUser(ctx context.Context, user PT) error {
	if err := s.prepareEntity(user); err != nil {
		return err
	}
	return s.userRep.Insert(ctx, user)
}

func (s *AuthService[PT, T]) UpdateUser(ctx context.Context, user PT) error {
	if err := s.prepareEntity(user); err != nil {
		return err
	}
	return s.userRep.Update(ctx, user)
}

func (s *AuthService[PT, T]) Login(ctx context.Context, username, password string) (PT, error) {
	user, err := s.userRep.FindByLogin(ctx, username, &repository.QueryOptions{Archived: false})
	if err != nil {
		return nil, err
	}

	if err := s.VerifyPassword(user.Password(), password); err != nil {
		return nil, ErrWrongPassword
	}
	return user, nil
}

func (s *AuthService[PT, T]) HashAndSalt(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func (s *AuthService[PT, T]) VerifyPassword(passwordHash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
}
