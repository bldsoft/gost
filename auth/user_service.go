package auth

import (
	"context"
)

// UserService ...
type UserService[PT IUserPtr[T], T any] struct {
	userRep        IUserRepository[PT]
	passwordHasher PasswordHasher
}

// NewUserService ...
func NewUserService[PT IUserPtr[T], T any](rep IUserRepository[PT], passwordHasher PasswordHasher) *UserService[PT, T] {
	return &UserService[PT, T]{userRep: rep, passwordHasher: passwordHasher}
}

func (s *UserService[PT, T]) UserFromContext(ctx context.Context, allowPanic bool) (PT, bool) {
	return FromContext[PT](ctx, allowPanic)
}

func (s *UserService[PT, T]) Create(ctx context.Context, user PT) error {
	hashedPass, err := s.passwordHasher.HashAndSalt(user.Password())
	if err != nil {
		return err
	}
	user.SetPassword(hashedPass)
	return s.userRep.Insert(ctx, user)
}

func (s *UserService[PT, T]) GetAll(ctx context.Context) ([]PT, error) {
	return s.userRep.GetAll(ctx)
}

func (s *UserService[PT, T]) GetByID(ctx context.Context, id string) (PT, error) {
	return s.userRep.FindByID(ctx, id)
}

func (s *UserService[PT, T]) Update(ctx context.Context, user PT) error {
	user.SetPassword("") // don't update password
	return s.userRep.Update(ctx, user)
}

func (s *UserService[PT, T]) UpdatePassword(ctx context.Context, id, password string) error {
	hashedPass, err := s.passwordHasher.HashAndSalt(password)
	if err != nil {
		return err
	}
	var user T
	PT(&user).SetIDFromString(id)
	PT(&user).SetPassword(hashedPass)
	return s.userRep.Update(ctx, &user)
}

func (s *UserService[PT, T]) Delete(ctx context.Context, id string) error {
	return s.userRep.Delete(ctx, id)
}