package auth

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var ErrWrongPassword = fmt.Errorf("wrong password")

// AuthService ...
type AuthService[T any, U AuthenticatablePtr[T]] struct {
	userRep IUserRepository[U]
}

// NewAuthService ...
func NewAuthService[T any, U AuthenticatablePtr[T]](rep IUserRepository[U]) *AuthService[T, U] {
	return &AuthService[T, U]{userRep: rep}
}

func (s *AuthService[T, U]) SignUp(ctx context.Context, user U) error {
	hashedPass, err := s.HashAndSalt(user.Password())
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}
	user.SetPassword(hashedPass)
	return s.userRep.Insert(ctx, user)
}

func (s *AuthService[T, U]) Login(ctx context.Context, username, password string) (U, error) {
	user, err := s.userRep.FindByName(ctx, username)
	if err != nil {
		return nil, err
	}

	if err := s.VerifyPassword(user.Password(), password); err != nil {
		return nil, ErrWrongPassword
	}
	return user, nil
}

func (s *AuthService[T, U]) HashAndSalt(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func (s *AuthService[T, U]) VerifyPassword(passwordHash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
}
