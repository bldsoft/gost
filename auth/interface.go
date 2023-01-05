package auth

import (
	"context"
	"errors"

	"github.com/bldsoft/gost/repository"
)

var ErrUnauthorized = errors.New("unauthorized")

type Authenticable interface {
	Login() string
	Password() string
	SetPassword(string)
}

type AuthenticablePtr[T any] interface {
	*T
	Authenticable
}

type IUserPtr[T any] interface {
	repository.IEntityID
	AuthenticablePtr[T]
}

type IRole interface {
	comparable
}

type PasswordHasher interface {
	HashAndSalt(password string) (string, error)
	VerifyPassword(passwordHash, password string) error
}

type IAuthRepository[PT any] interface {
	FindByLogin(ctx context.Context, login string, options ...*repository.QueryOptions[PT]) (PT, error)
}

// IAuthService ...
type IAuthService[PT AuthenticablePtr[T], T any] interface {
	Login(ctx context.Context, username, password string) (PT, error)
}

type IUserRepository[PT any] interface {
	Insert(ctx context.Context, user PT) error
	GetAll(ctx context.Context, options ...*repository.QueryOptions[PT]) ([]PT, error)
	FindByID(ctx context.Context, id interface{}, options ...*repository.QueryOptions[PT]) (PT, error)
	Update(ctx context.Context, user PT) error
	Delete(ctx context.Context, id interface{}, options ...*repository.QueryOptions[PT]) error
}

type IUserService[PT AuthenticablePtr[T], T any] interface {
	Create(ctx context.Context, user PT) error
	GetAll(ctx context.Context, archived bool) ([]PT, error)
	GetByID(ctx context.Context, id string) (PT, error)
	Update(ctx context.Context, user PT) error
	UpdatePassword(ctx context.Context, id, password string) error
	Delete(ctx context.Context, id string, archived bool) error
}
