package auth

import (
	"context"
	"fmt"

	"github.com/bldsoft/gost/repository"
	"golang.org/x/crypto/bcrypt"
)

var ErrWrongPassword = fmt.Errorf("wrong password")

// AuthService ...
type AuthService[U AuthenticatablePtr[T], T any] struct {
	userRep IUserRepository[U]
}

// NewAuthService ...
func NewAuthService[U AuthenticatablePtr[T], T any](rep IUserRepository[U]) *AuthService[U, T] {
	return &AuthService[U, T]{userRep: rep}
}

func (s *AuthService[U, T]) UserFromContext(ctx context.Context, allowPanic bool) (U, bool) {
	return FromContext[U](ctx, allowPanic)
}

func (s *AuthService[U, T]) prepareEntity(user U) error {
	hashedPass, err := s.HashAndSalt(user.Password())
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}
	user.SetPassword(hashedPass)
	return nil
}

func (s *AuthService[U, T]) CreateUser(ctx context.Context, user U) error {
	if err := s.prepareEntity(user); err != nil {
		return err
	}
	return s.userRep.Insert(ctx, user)
}

func (s *AuthService[U, T]) UpdateUser(ctx context.Context, user U) error {
	if err := s.prepareEntity(user); err != nil {
		return err
	}
	return s.userRep.Update(ctx, user)
}

func (s *AuthService[U, T]) Login(ctx context.Context, username, password string) (U, error) {
	user, err := s.userRep.FindByName(ctx, username, &repository.QueryOptions{Archived: false})
	if err != nil {
		return nil, err
	}

	if err := s.VerifyPassword(user.Password(), password); err != nil {
		return nil, ErrWrongPassword
	}
	return user, nil
}

func (s *AuthService[U, T]) HashAndSalt(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func (s *AuthService[U, T]) VerifyPassword(passwordHash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
}
