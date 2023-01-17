package mongo

import (
	"context"

	"github.com/bldsoft/gost/repository"
	"go.mongodb.org/mongo-driver/mongo"
)

type Repository[T any, U repository.IEntityIDPtr[T]] interface {
	Name() string
	Collection() *mongo.Collection
	WithTransaction(ctx context.Context, f func(ctx mongo.SessionContext) (interface{}, error)) (interface{}, error)

	FindOne(ctx context.Context, filter interface{}, opts ...*repository.QueryOptions) (U, error)
	FindByID(ctx context.Context, id interface{}, options ...*repository.QueryOptions) (U, error)
	FindByStringIDs(ctx context.Context, ids []string, preserveOrder bool, options ...*repository.QueryOptions) ([]U, error)
	FindByIDs(ctx context.Context, ids []interface{}, preserveOrder bool, options ...*repository.QueryOptions) ([]U, error)
	Find(ctx context.Context, filter interface{}, opt ...*repository.QueryOptions) ([]U, error)
	GetAll(ctx context.Context, options ...*repository.QueryOptions) ([]U, error)

	Insert(ctx context.Context, entity U) error
	InsertMany(ctx context.Context, entities []U) error
	Update(ctx context.Context, entity U, options ...*repository.QueryOptions) error
	UpdateMany(ctx context.Context, entities []U) error
	UpdateOne(ctx context.Context, filter interface{}, update interface{}, options ...*repository.QueryOptions) error
	UpdateAndGetByID(ctx context.Context, updateEntity U, returnNewDocument bool, queryOpt ...*repository.QueryOptions) (U, error)
	UpsertOne(ctx context.Context, filter interface{}, update U) error
	Delete(ctx context.Context, id interface{}, options ...*repository.QueryOptions) error
	DeleteMany(ctx context.Context, filter interface{}, options ...*repository.QueryOptions) error

	AggregateOne(ctx context.Context, pipeline mongo.Pipeline, entity interface{}) error
}

type IFilter interface {
	Filter(f interface{})
}
