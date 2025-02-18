package auth

import (
	"context"

	"github.com/bldsoft/gost/repository"
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

func (s *UserService[PT, T]) Create(ctx context.Context, user PT, recoverDeleted bool) error {
	hashedPass, err := s.passwordHasher.HashAndSalt(user.Password())
	if err != nil {
		return err
	}
	user.SetPassword(hashedPass)

	if recoverDeleted {
		return s.userRep.InsertOrRecover(ctx, user)
	}
	return s.userRep.Insert(ctx, user)
}

func (s *UserService[PT, T]) GetAll(ctx context.Context, archived bool) ([]PT, error) {
	return s.userRep.GetAll(ctx, &repository.QueryOptions{Archived: archived})
}

func (s *UserService[PT, T]) GetByID(ctx context.Context, id string) (PT, error) {
	return s.userRep.FindByID(ctx, id)
}

func (s *UserService[PT, T]) Update(ctx context.Context, user PT) error {
	if password := user.Password(); password != "" {
		hashedPass, err := s.passwordHasher.HashAndSalt(password)
		if err != nil {
			return err
		}
		user.SetPassword(hashedPass)
	}
	return s.userRep.Update(ctx, user)
}

func (s *UserService[PT, T]) UpdatePassword(ctx context.Context, id, password string) error {
	hashedPass, err := s.passwordHasher.HashAndSalt(password)
	if err != nil {
		return err
	}
	user, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	user.SetPassword(hashedPass)
	return s.userRep.Update(ctx, user)
}

func (s *UserService[PT, T]) Delete(ctx context.Context, id string, archived bool) error {
	return s.userRep.Delete(ctx, id, &repository.QueryOptions{Archived: archived})
}
