package auth

import (
	"context"

	"github.com/bldsoft/gost/repository"
)

type Authenticatable interface {
	Name() string
	Password() string
	SetPassword(string)
}

type AuthenticatablePtr[T any] interface {
	*T
	Authenticatable
}

type IRole interface {
	comparable
}

type Authorizable[T IRole] interface {
	Role() T
}

type IUserRepository[U any] interface {
	Update(ctx context.Context, user U) error
	Insert(ctx context.Context, user U) error
	FindByName(ctx context.Context, name string, options ...*repository.QueryOptions) (U, error)
}

type PasswordHasher interface {
	HashAndSalt(password string) (string, error)
	VerifyPassword(passwordHash, password string) error
}

// IAuthService ...
type IAuthService[U AuthenticatablePtr[T], T any] interface {
	PasswordHasher
	CreateUser(ctx context.Context, user U) error
	UpdateUser(ctx context.Context, user U) error
	Login(ctx context.Context, username, password string) (U, error)
}
