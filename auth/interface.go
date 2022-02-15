package auth

import (
	"context"

	"github.com/bldsoft/gost/mongo"
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

type IUserPtr[T any] interface {
	mongo.IEntityID
	AuthenticatablePtr[T]
}

type IRole interface {
	comparable
}

type Authorizable[T IRole] interface {
	Role() T
}
type PasswordHasher interface {
	HashAndSalt(password string) (string, error)
	VerifyPassword(passwordHash, password string) error
}

type IAuthRepository[PT any] interface {
	FindByLogin(ctx context.Context, login string, options ...*repository.QueryOptions) (PT, error)
}

// IAuthService ...
type IAuthService[PT AuthenticatablePtr[T], T any] interface {
	Login(ctx context.Context, username, password string) (PT, error)
}

type IUserRepository[PT any] interface {
	Insert(ctx context.Context, user PT) error 
	GetAll(ctx context.Context,options ...*repository.QueryOptions) ([]PT, error) 
	FindByID(ctx context.Context, id string,options ...*repository.QueryOptions) (PT, error)
	Update(ctx context.Context, user PT) error
	Delete(ctx context.Context, user PT, options ...*repository.QueryOptions) error
}


type IUserService[PT AuthenticatablePtr[T], T any] interface {
	Create(ctx context.Context, user PT) error 
	GetAll(ctx context.Context) ([]PT, error) 
	GetByID(ctx context.Context, id string) (PT, error)
	Update(ctx context.Context, user PT) error
	Delete(ctx context.Context, user PT) error
}