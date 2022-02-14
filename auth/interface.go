package auth

import (
	"context"

	"github.com/bldsoft/gost/repository"
)

type Authenticatable interface {
	Login() string
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

type IUserRepository[PT any] interface {
	Update(ctx context.Context, user PT) error
	Insert(ctx context.Context, user PT) error
	FindByLogin(ctx context.Context, name string, options ...*repository.QueryOptions) (PT, error)
}

type PasswordHasher interface {
	HashAndSalt(password string) (string, error)
	VerifyPassword(passwordHash, password string) error
}

// IAuthService ...
type IAuthService[PT AuthenticatablePtr[T], T any] interface {
	PasswordHasher
	CreateUser(ctx context.Context, user PT) error
	UpdateUser(ctx context.Context, user PT) error
	Login(ctx context.Context, username, password string) (PT, error)
}
